//go:generate go run tools/genprefixes.go

package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/fatih/color"
)

type Result struct {
	i        int
	s        string
	matched  bool
	ordinary bool
}

type Target struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

var defaultTargets = []Target{
	{Name: "北京电信", IP: "219.141.140.10"},
	{Name: "北京联通", IP: "202.106.195.68"},
	{Name: "北京移动", IP: "221.179.155.161"},
	{Name: "上海电信", IP: "202.96.209.133"},
	{Name: "上海联通", IP: "210.22.97.1"},
	{Name: "上海移动", IP: "211.136.112.200"},
	{Name: "广州电信", IP: "58.60.188.222"},
	{Name: "广州联通", IP: "210.21.196.6"},
	{Name: "广州移动", IP: "120.196.165.24"},
	{Name: "枣庄电信", IP: "219.146.1.66"},
	{Name: "枣庄联通", IP: "202.102.128.68"},
	{Name: "枣庄移动", IP: "218.201.96.130"},
	{Name: "济宁电信", IP: "222.175.169.91"},
	{Name: "济宁联通", IP: "202.102.154.3"},
	{Name: "济宁移动", IP: "211.137.191.26"},
	{Name: "吉林电信", IP: "222.168.134.149"},
	{Name: "吉林联通", IP: "202.98.0.68"},
	{Name: "吉林移动", IP: "211.141.16.99"},
}

var routeLabels = map[string]string{
	"AS4809":  "电信CN2 GIA [顶级线路]",
	"AS9929":  "联通9929 [顶级线路]",
	"AS58807": "移动CMIN2 [顶级线路]",
}

var asnPrefixes = func() []struct {
	asn  string
	cidr *net.IPNet
} {
	out := make([]struct {
		asn  string
		cidr *net.IPNet
	}, 0, len(asnPrefixStrings))
	for _, p := range asnPrefixStrings {
		_, c, err := net.ParseCIDR(p.cidr)
		if err != nil {
			panic(err)
		}
		out = append(out, struct {
			asn  string
			cidr *net.IPNet
		}{p.asn, c})
	}
	return out
}()

func trace(ch chan Result, i int, target Target) {
	ch <- traceTarget(i, target)
}

func traceTarget(i int, target Target) Result {
	hops, err := Trace(net.ParseIP(target.IP))
	if err != nil {
		return Result{i: i, s: formatResult(target, fmt.Sprint(err))}
	}

	for _, h := range hops {
		for _, n := range h.Nodes {
			asn := ipAsn(n.IP.String())
			if asn == "" {
				continue
			}
			label := routeLabels[asn]
			if label == "" {
				continue
			}
			return Result{i: i, s: formatResult(target, formatRouteHit(asn, label, n.IP.String())), matched: true}
		}
	}

	if len(hops) == 0 {
		return Result{i: i, s: formatResult(target, color.New(color.FgRed).Add(color.Bold).Sprint("超时"))}
	}
	return Result{i: i, s: formatResult(target, color.New(color.FgWhite).Sprint("普通线路")), ordinary: true}
}

func retryOrdinaryResults(results []Result, targets []Target) {
	for i := range results {
		if !results[i].ordinary {
			continue
		}
		for attempt := 0; attempt < ordinaryRetryCount; attempt++ {
			retry := traceTarget(i, targets[i])
			if retry.matched {
				results[i] = retry
				break
			}
		}
	}
}

func formatResult(target Target, route string) string {
	return padRight(target.Name, 12) + padRight(target.IP, 17) + route
}

func formatRouteHit(asn, label, hop string) string {
	return colorizeRoute(asn, label) + " 命中 " + hop
}

func padRight(s string, width int) string {
	w := displayWidth(s)
	if w >= width {
		return s + " "
	}
	return s + strings.Repeat(" ", width-w)
}

func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r >= 0x2E80 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

func colorizeRoute(asn, label string) string {
	switch asn {
	case "AS9929":
		return color.New(color.FgHiYellow).Add(color.Bold).Sprint(label)
	case "AS4809":
		return color.New(color.FgHiMagenta).Add(color.Bold).Sprint(label)
	case "AS58807":
		return color.New(color.FgHiBlue).Add(color.Bold).Sprint(label)
	}
	return label
}

func ipAsn(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ""
	}
	for _, p := range asnPrefixes {
		if p.cidr.Contains(parsed) {
			return p.asn
		}
	}
	return ""
}
