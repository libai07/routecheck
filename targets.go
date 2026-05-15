package main

import (
	"fmt"
	"net"
	"strings"

	"github.com/fatih/color"
)

type Result struct {
	index int
	line  string
}

type Target struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

const (
	targetNameColumnWidth = 12
	targetIPColumnWidth   = 17
	routeLabelColumnWidth = 24
)

type routePrefix struct {
	asn   string
	label string
	cidr  string
}

var defaultTargets = []Target{
	{Name: "北京电信", IP: "219.141.140.10"},
	{Name: "北京联通", IP: "202.106.195.68"},
	{Name: "北京移动", IP: "221.130.33.60"},
	{Name: "上海电信", IP: "202.96.209.133"},
	{Name: "上海联通", IP: "210.22.70.3"},
	{Name: "上海移动", IP: "211.136.112.50"},
	{Name: "天津电信", IP: "219.150.32.132"},
	{Name: "天津联通", IP: "202.99.104.68"},
	{Name: "天津移动", IP: "211.137.160.50"},
	{Name: "重庆电信", IP: "61.128.192.68"},
	{Name: "重庆联通", IP: "221.5.203.98"},
	{Name: "重庆移动", IP: "218.201.4.3"},
}

var premiumPrefixStrings = []routePrefix{
	{"AS4809", "电信CN2 GIA [顶级线路]", "59.43.0.0/16"},
	{"AS9929", "联通9929 [顶级线路]", "218.105.0.0/16"},
	{"AS58807", "移动CMIN2 [顶级线路]", "223.120.128.0/17"},
}

var ordinaryPrefixStrings = []routePrefix{
	{"", "电信163 [普通线路]", "202.97.0.0/16"},
	{"", "联通4837 [普通线路]", "219.158.0.0/16"},
	{"", "移动CMI [普通线路]", "223.120.0.0/17"},
	{"", "移动CMNET [普通线路]", "221.183.0.0/16"},
}

var premiumPrefixes = parseRoutePrefixes(premiumPrefixStrings)
var ordinaryPrefixes = parseRoutePrefixes(ordinaryPrefixStrings)

func trace(ch chan Result, i int, target Target) {
	ch <- traceTarget(i, target)
}

func traceTarget(i int, target Target) Result {
	ips, err := traceRoute(target.IP)
	if err != nil {
		return Result{index: i, line: formatResult(target, formatRouteOnly(fmt.Sprint(err), color.New(color.FgRed).SprintFunc()))}
	}

	for _, ip := range ips {
		asn, label := premiumRoute(ip.String())
		if label == "" {
			continue
		}
		return Result{index: i, line: formatResult(target, formatRouteHit(asn, label, ip.String()))}
	}

	if len(ips) == 0 {
		return Result{index: i, line: formatResult(target, formatRouteOnly("超时", color.New(color.FgRed).Add(color.Bold).SprintFunc()))}
	}

	for _, ip := range ips {
		label := ordinaryRoute(ip.String())
		if label == "" {
			continue
		}
		return Result{index: i, line: formatResult(target, formatOrdinaryHit(label, ip.String()))}
	}

	return Result{index: i, line: formatResult(target, formatRouteOnly("未命中精品线路", color.New(color.FgWhite).SprintFunc()))}
}

func formatResult(target Target, route string) string {
	return padRight(target.Name, targetNameColumnWidth) + padRight(target.IP, targetIPColumnWidth) + route
}

func formatRouteHit(asn, label, hop string) string {
	return colorizeRoute(asn, padRight(label, routeLabelColumnWidth)) + "命中 " + hop
}

func formatOrdinaryHit(label, hop string) string {
	return color.New(color.FgWhite).Sprint(padRight(label, routeLabelColumnWidth)) + "命中 " + hop
}

func formatRouteOnly(label string, colorize func(a ...interface{}) string) string {
	return colorize(padRight(label, routeLabelColumnWidth))
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

func ordinaryRoute(ip string) string {
	_, label := matchRoutePrefix(ip, ordinaryPrefixes)
	return label
}

func premiumRoute(ip string) (string, string) {
	return matchRoutePrefix(ip, premiumPrefixes)
}

func matchRoutePrefix(ip string, prefixes []struct {
	asn   string
	label string
	cidr  *net.IPNet
}) (string, string) {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", ""
	}
	for _, p := range prefixes {
		if p.cidr.Contains(parsed) {
			return p.asn, p.label
		}
	}
	return "", ""
}

func parseRoutePrefixes(prefixes []routePrefix) []struct {
	asn   string
	label string
	cidr  *net.IPNet
} {
	out := make([]struct {
		asn   string
		label string
		cidr  *net.IPNet
	}, 0, len(prefixes))
	for _, p := range prefixes {
		_, c, err := net.ParseCIDR(p.cidr)
		if err != nil {
			panic(err)
		}
		out = append(out, struct {
			asn   string
			label string
			cidr  *net.IPNet
		}{p.asn, p.label, c})
	}
	return out
}
