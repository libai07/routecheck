package main

import "testing"

func TestIPAsn(t *testing.T) {
	cases := []struct {
		ip   string
		want string
		note string
	}{
		{"59.43.1.1", "AS4809", "CN2 GIA 核心段"},
		{"59.43.188.1", "AS4809", "CN2 GIA 典型 hop"},
		{"218.105.1.1", "AS9929", "联通 9929 经典段"},
		{"210.51.1.1", "AS9929", "联通 9929 上海段"},
		{"223.120.128.1", "AS58807", "CMIN2 核心 /17"},
		{"223.120.200.1", "AS58807", "CMIN2 上半段"},
		{"223.119.8.1", "AS58807", "CMIN2 国际段"},

		{"8.8.8.8", "", "Google DNS，不在三网顶级线路"},
		{"202.97.1.1", "", "AS4134 电信 163 普通线路"},
		{"219.158.1.1", "", "AS4837 联通 4837 普通线路"},
		{"223.120.16.1", "", "AS58453 CMI 普通线路，不是 CMIN2"},
		{"1.2.3.4", "", "随机非相关 IP"},
		{"not-an-ip", "", "无效输入"},
	}

	for _, c := range cases {
		got := ipAsn(c.ip)
		if got != c.want {
			t.Errorf("ipAsn(%q) = %q, want %q [%s]", c.ip, got, c.want, c.note)
		}
	}
}

func TestAsnPrefixesLoaded(t *testing.T) {
	if len(asnPrefixes) == 0 {
		t.Fatal("asnPrefixes is empty; generated prefixes not loaded")
	}
	if len(asnPrefixStrings) == 0 {
		t.Fatal("asnPrefixStrings is empty; prefixes_generated.go not compiled in")
	}
	if len(asnPrefixes) != len(asnPrefixStrings) {
		t.Errorf("len mismatch: asnPrefixes=%d, asnPrefixStrings=%d",
			len(asnPrefixes), len(asnPrefixStrings))
	}
	t.Logf("loaded %d CIDR prefixes", len(asnPrefixes))

	counts := map[string]int{}
	for _, p := range asnPrefixStrings {
		counts[p.asn]++
	}
	for _, asn := range []string{"AS4809", "AS9929", "AS58807"} {
		if counts[asn] == 0 {
			t.Errorf("no prefixes for %s", asn)
		}
		t.Logf("%s: %d prefixes", asn, counts[asn])
	}
}
