//go:build !linux

package main

import (
	"context"
	"errors"
	"net"
	"time"
)

var errUnsupportedPlatform = errors.New("traceroute only supports Linux")

type Reply struct {
	IP   net.IP
	RTT  time.Duration
	Hops int
}

type Node struct {
	IP  net.IP
	RTT []time.Duration
}

type Hop struct {
	Nodes    []*Node
	Distance int
}

func Trace(ip net.IP) ([]*Hop, error) {
	return nil, errUnsupportedPlatform
}

type Tracer struct{}

func (t *Tracer) Trace(ctx context.Context, ip net.IP, h func(reply *Reply)) error {
	return errUnsupportedPlatform
}
