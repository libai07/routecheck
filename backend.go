package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func traceRoute(ip string) ([]net.IP, error) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return nil, fmt.Errorf("无效 IP: %s", ip)
	}
	return traceNextTrace(parsed.String())
}

func traceNextTrace(ip string) ([]net.IP, error) {
	bin, err := findNextTraceBinary()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), traceTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, "--json", "--no-color", "--no-rdns", "--data-provider", "disable-geoip", ip)
	out, err := cmd.Output()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && len(exitErr.Stderr) > 0 {
			return nil, fmt.Errorf("NextTrace 执行失败: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	return parseNextTraceIPs(out)
}

func findNextTraceBinary() (string, error) {
	for _, name := range []string{"nexttrace", "nexttrace-tiny"} {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}
	return "", errors.New("未找到 nexttrace 或 nexttrace-tiny")
}

func parseNextTraceIPs(data []byte) ([]net.IP, error) {
	data = trimToJSONObject(data)
	if len(data) == 0 {
		return nil, errors.New("nexttrace JSON 为空")
	}

	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}

	rawHops, ok := lookupMapKey(root, "hops").([]any)
	if !ok {
		return nil, errors.New("nexttrace JSON 缺少 Hops")
	}

	ips := make([]net.IP, 0, len(rawHops))
	for _, rawTTL := range rawHops {
		ips = append(ips, extractTTLIPs(rawTTL)...)
	}
	return ips, nil
}

func trimToJSONObject(data []byte) []byte {
	data = bytes.TrimSpace(data)
	start := bytes.IndexByte(data, '{')
	end := bytes.LastIndexByte(data, '}')
	if start < 0 || end < start {
		return nil
	}
	return data[start : end+1]
}

func extractTTLIPs(raw any) []net.IP {
	items, ok := raw.([]any)
	if !ok {
		items = []any{raw}
	}

	seen := map[string]bool{}
	out := make([]net.IP, 0, len(items))
	for _, item := range items {
		for _, ip := range extractIPsFromItem(item) {
			key := ip.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, ip)
		}
	}
	return out
}

func extractIPsFromItem(raw any) []net.IP {
	obj, ok := raw.(map[string]any)
	if !ok {
		if ip := parseIPFromValue(raw); ip != nil {
			return []net.IP{ip}
		}
		return nil
	}

	out := []net.IP{}
	for _, key := range []string{"attempts", "nodes", "probes", "responses", "results"} {
		if child, ok := lookupMapKey(obj, key).([]any); ok {
			for _, item := range child {
				out = append(out, extractIPsFromItem(item)...)
			}
		}
	}
	if ip := extractHopIP(obj); ip != nil {
		out = append(out, ip)
	}
	return out
}

func extractHopIP(raw any) net.IP {
	obj, ok := raw.(map[string]any)
	if !ok {
		return parseIPFromValue(raw)
	}

	for _, key := range []string{"address", "ip", "addr", "resolvedaddress"} {
		if ip := parseIPFromValue(lookupMapKey(obj, key)); ip != nil {
			return ip
		}
	}
	return nil
}

func parseIPFromValue(v any) net.IP {
	switch x := v.(type) {
	case string:
		return parseLooseIP(x)
	case map[string]any:
		for _, key := range []string{"ip", "address", "addr"} {
			if ip := parseIPFromValue(lookupMapKey(x, key)); ip != nil {
				return ip
			}
		}
	case []any:
		for _, item := range x {
			if ip := parseIPFromValue(item); ip != nil {
				return ip
			}
		}
	}
	return nil
}

func parseLooseIP(s string) net.IP {
	s = strings.TrimSpace(s)
	if s == "" || s == "*" {
		return nil
	}
	if ip := net.ParseIP(s); ip != nil {
		return ip
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		return net.ParseIP(host)
	}
	return nil
}

func lookupMapKey(m map[string]any, key string) any {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return nil
}
