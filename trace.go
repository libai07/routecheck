//go:build linux

package main

import (
	"context"
	"errors"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
)

var DefaultTracer = &Tracer{Config: Config{
	Delay:        75 * time.Millisecond,
	Timeout:      1500 * time.Millisecond,
	MaxHops:      30,
	ProbesPerHop: 3,
	Networks:     []string{"ip4:icmp", "ip4:ip"},
}}

type Config struct {
	Delay        time.Duration
	Timeout      time.Duration
	MaxHops      int
	ProbesPerHop int
	Networks     []string
}

type Tracer struct {
	Config

	once sync.Once
	conn *net.IPConn
	err  error

	mu   sync.RWMutex
	sess map[string][]*Session
	seq  uint32
}

func (t *Tracer) Trace(ctx context.Context, ip net.IP, h func(reply *Reply)) error {
	sess, err := t.NewSession(ip)
	if err != nil {
		return err
	}
	defer sess.Close()

	delay := time.NewTicker(t.Delay)
	defer delay.Stop()

	max := t.MaxHops
	for ttl := 1; ttl <= t.MaxHops && ttl <= max; ttl++ {
		for probe := 0; probe < t.ProbesPerHop; probe++ {
			err = sess.Ping(ttl)
			if err != nil {
				return err
			}
			select {
			case <-delay.C:
			case r := <-sess.Receive():
				if max > r.Hops && ip.Equal(r.IP) {
					max = r.Hops
				}
				h(r)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
	if sess.isDone(max) {
		return nil
	}

	deadline := time.After(t.Timeout)
	for {
		select {
		case r := <-sess.Receive():
			if max > r.Hops && ip.Equal(r.IP) {
				max = r.Hops
			}
			h(r)
			if sess.isDone(max) {
				return nil
			}
		case <-deadline:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (t *Tracer) NewSession(ip net.IP) (*Session, error) {
	t.once.Do(t.init)
	if t.err != nil {
		return nil, t.err
	}
	return newSession(t, shortIP(ip)), nil
}

func (t *Tracer) init() {
	for _, network := range t.Networks {
		t.conn, t.err = t.listen(network)
		if t.err != nil {
			continue
		}
		go t.serve(t.conn)
		return
	}
}

func (t *Tracer) listen(network string) (*net.IPConn, error) {
	conn, err := net.ListenIP(network, nil)
	if err != nil {
		return nil, err
	}
	raw, err := conn.SyscallConn()
	if err != nil {
		conn.Close()
		return nil, err
	}
	_ = raw.Control(func(fd uintptr) {
		err = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_HDRINCL, 1)
	})
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

func (t *Tracer) serve(conn *net.IPConn) error {
	defer conn.Close()
	buf := make([]byte, 1500)
	for {
		n, from, err := conn.ReadFromIP(buf)
		if err != nil {
			return err
		}
		if err := t.serveData(from.IP, buf[:n]); err != nil {
			continue
		}
	}
}

func (t *Tracer) serveData(from net.IP, b []byte) error {
	if from.To4() == nil {
		return errUnsupportedProtocol
	}
	now := time.Now()
	msg, err := icmp.ParseMessage(ProtocolICMP, b)
	if err != nil {
		return err
	}
	if msg.Type == ipv4.ICMPTypeEchoReply {
		echo := msg.Body.(*icmp.Echo)
		return t.serveReply(from, &packet{from, uint16(echo.ID), 1, now})
	}

	b = getReplyData(msg)
	if len(b) < ipv4.HeaderLen {
		return errMessageTooShort
	}

	ip, err := ipv4.ParseHeader(b)
	if err != nil {
		return err
	}
	return t.serveReply(ip.Dst, &packet{from, uint16(ip.ID), ip.TTL, now})
}

func (t *Tracer) sendRequest(dst net.IP, ttl int) (*packet, error) {
	id := uint16(atomic.AddUint32(&t.seq, 1))
	b := newPacket(id, dst, ttl)
	req := &packet{dst, id, ttl, time.Now()}
	_, err := t.conn.WriteToIP(b, &net.IPAddr{IP: dst})
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (t *Tracer) addSession(s *Session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.sess == nil {
		t.sess = make(map[string][]*Session)
	}
	t.sess[string(s.ip)] = append(t.sess[string(s.ip)], s)
}

func (t *Tracer) removeSession(s *Session) {
	t.mu.Lock()
	defer t.mu.Unlock()
	a := t.sess[string(s.ip)]
	for i, it := range a {
		if it == s {
			t.sess[string(s.ip)] = append(a[:i], a[i+1:]...)
			return
		}
	}
}

func (t *Tracer) serveReply(dst net.IP, res *packet) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	a := t.sess[string(shortIP(dst))]
	for _, s := range a {
		s.handle(res)
	}
	return nil
}

type Session struct {
	t  *Tracer
	ip net.IP
	ch chan *Reply

	mu     sync.RWMutex
	probes []*packet
}

func newSession(t *Tracer, ip net.IP) *Session {
	s := &Session{
		t:  t,
		ip: ip,
		ch: make(chan *Reply, 64),
	}
	t.addSession(s)
	return s
}

func (s *Session) Ping(ttl int) error {
	req, err := s.t.sendRequest(s.ip, ttl+1)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.probes = append(s.probes, req)
	s.mu.Unlock()
	return nil
}

func (s *Session) Receive() <-chan *Reply {
	return s.ch
}

func (s *Session) isDone(ttl int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.probes {
		if r.TTL <= ttl {
			return false
		}
	}
	return true
}

func (s *Session) handle(res *packet) {
	now := res.Time
	n := 0
	var req *packet
	s.mu.Lock()
	for _, r := range s.probes {
		if now.Sub(r.Time) > s.t.Timeout {
			continue
		}
		if r.ID == res.ID {
			req = r
			continue
		}
		s.probes[n] = r
		n++
	}
	s.probes = s.probes[:n]
	s.mu.Unlock()
	if req == nil {
		return
	}

	hops := req.TTL - res.TTL + 1
	if hops < 1 {
		hops = 1
	}
	select {
	case s.ch <- &Reply{IP: res.IP, Hops: hops}:
	default:
	}
}

func (s *Session) Close() {
	s.t.removeSession(s)
}

type packet struct {
	IP   net.IP
	ID   uint16
	TTL  int
	Time time.Time
}

func shortIP(ip net.IP) net.IP {
	if v := ip.To4(); v != nil {
		return v
	}
	return ip
}

func getReplyData(msg *icmp.Message) []byte {
	switch b := msg.Body.(type) {
	case *icmp.TimeExceeded:
		return b.Data
	case *icmp.DstUnreach:
		return b.Data
	case *icmp.ParamProb:
		return b.Data
	}
	return nil
}

var (
	errMessageTooShort     = errors.New("message too short")
	errUnsupportedProtocol = errors.New("unsupported protocol")
)

func newPacket(id uint16, dst net.IP, ttl int) []byte {
	msg := icmp.Message{
		Type: ipv4.ICMPTypeEcho,
		Body: &icmp.Echo{
			ID:  int(id),
			Seq: int(id),
		},
	}
	p, _ := msg.Marshal(nil)
	ip := &ipv4.Header{
		Version:  ipv4.Version,
		Len:      ipv4.HeaderLen,
		TotalLen: ipv4.HeaderLen + len(p),
		TOS:      16,
		ID:       int(id),
		Dst:      dst,
		Protocol: ProtocolICMP,
		TTL:      ttl,
	}
	buf, err := ip.Marshal()
	if err != nil {
		return nil
	}
	return append(buf, p...)
}

const ProtocolICMP = 1

type Reply struct {
	IP   net.IP
	Hops int
}

type Node struct {
	IP net.IP
}

type Hop struct {
	Nodes    []*Node
	Distance int
}

func (h *Hop) Add(r *Reply) {
	for _, it := range h.Nodes {
		if it.IP.Equal(r.IP) {
			return
		}
	}
	h.Nodes = append(h.Nodes, &Node{IP: r.IP})
}

func Trace(ip net.IP) ([]*Hop, error) {
	hops := make([]*Hop, 0, DefaultTracer.MaxHops)
	touch := func(dist int) *Hop {
		for _, h := range hops {
			if h.Distance == dist {
				return h
			}
		}
		h := &Hop{Distance: dist}
		hops = append(hops, h)
		return h
	}
	err := DefaultTracer.Trace(context.Background(), ip, func(r *Reply) {
		touch(r.Hops).Add(r)
	})
	if err != nil && err != context.DeadlineExceeded {
		return nil, err
	}
	sort.Slice(hops, func(i, j int) bool {
		return hops[i].Distance < hops[j].Distance
	})
	return hops, nil
}
