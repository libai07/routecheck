package main

import (
	"strings"
	"testing"

	"github.com/fatih/color"
)

func TestOrdinaryRoute(t *testing.T) {
	cases := []struct {
		ip   string
		want string
	}{
		{"202.97.1.1", "电信163 [普通线路]"},
		{"219.158.1.1", "联通4837 [普通线路]"},
		{"223.120.16.1", "移动CMI [普通线路]"},
		{"221.183.1.1", "移动CMNET [普通线路]"},
		{"218.105.1.1", ""},
		{"8.8.8.8", ""},
		{"not-an-ip", ""},
	}

	for _, c := range cases {
		got := ordinaryRoute(c.ip)
		if got != c.want {
			t.Errorf("ordinaryRoute(%q) = %q, want %q", c.ip, got, c.want)
		}
	}
}

func TestPremiumRoute(t *testing.T) {
	cases := []struct {
		ip        string
		wantASN   string
		wantLabel string
	}{
		{"59.43.188.1", "AS4809", "电信CN2 GIA [顶级线路]"},
		{"218.105.2.89", "AS9929", "联通9929 [顶级线路]"},
		{"223.120.201.69", "AS58807", "移动CMIN2 [顶级线路]"},
		{"210.51.1.1", "", ""},
		{"223.119.8.1", "", ""},
		{"202.97.1.1", "", ""},
		{"not-an-ip", "", ""},
	}

	for _, c := range cases {
		gotASN, gotLabel := premiumRoute(c.ip)
		if gotASN != c.wantASN || gotLabel != c.wantLabel {
			t.Errorf("premiumRoute(%q) = (%q, %q), want (%q, %q)",
				c.ip, gotASN, gotLabel, c.wantASN, c.wantLabel)
		}
	}
}

func TestFormatRouteHitAlignment(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = oldNoColor
	}()

	results := []string{
		formatRouteHit("AS4809", "电信CN2 GIA [顶级线路]", "59.43.39.177"),
		formatRouteHit("AS9929", "联通9929 [顶级线路]", "218.105.2.89"),
		formatRouteHit("AS58807", "移动CMIN2 [顶级线路]", "223.120.201.69"),
		formatOrdinaryHit("电信163 [普通线路]", "202.97.1.1"),
		formatOrdinaryHit("联通4837 [普通线路]", "219.158.1.1"),
		formatOrdinaryHit("移动CMNET [普通线路]", "221.183.1.1"),
	}

	want := -1
	for _, result := range results {
		i := strings.Index(result, "命中")
		if i < 0 {
			t.Fatalf("result %q does not contain 命中", result)
		}
		got := displayWidth(result[:i])
		if want < 0 {
			want = got
		}
		if got != want {
			t.Fatalf("命中 column width = %d, want %d in %q", got, want, result)
		}
	}
}

func TestFormatRouteOnlyWidth(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() {
		color.NoColor = oldNoColor
	}()

	got := displayWidth(formatRouteOnly("未命中精品线路", color.New(color.FgWhite).SprintFunc()))
	want := routeLabelColumnWidth
	if got != want {
		t.Fatalf("route-only width = %d, want %d", got, want)
	}
}
