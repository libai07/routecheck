package main

import "testing"

func TestParseNextTraceIPs(t *testing.T) {
	data := []byte(`noise
{
  "Hops": [
    [
      {"Success": true, "Address": {"IP": "192.0.2.1", "Zone": ""}},
      {"Success": true, "Address": "192.0.2.1"}
    ],
    [
      {"success": true, "address": "218.105.2.89"}
    ],
    [
      {"success": false, "address": null}
    ],
    [
      {"ip": "223.120.201.69:0"}
    ]
  ]
}
tail`)

	ips, err := parseNextTraceIPs(data)
	if err != nil {
		t.Fatalf("parseNextTraceIPs error = %v", err)
	}
	if len(ips) != 3 {
		t.Fatalf("len(ips) = %d, want 3", len(ips))
	}
	if got := ips[0].String(); got != "192.0.2.1" {
		t.Fatalf("first hop = %s", got)
	}
	if got := ips[1].String(); got != "218.105.2.89" {
		t.Fatalf("second hop = %s", got)
	}
	if got := ips[2].String(); got != "223.120.201.69" {
		t.Fatalf("third hop = %s", got)
	}
}

func TestParseNextTraceIPsWithAttempts(t *testing.T) {
	data := []byte(`{
  "hops": [
    {
      "ttl": 1,
      "attempts": [
        {"success": true, "ip": "192.0.2.1"},
        {"success": true, "address": "192.0.2.1"}
      ]
    },
    {
      "ttl": 2,
      "attempts": [
        {"success": false},
        {"success": true, "ip": "218.105.2.89"}
      ]
    },
    {
      "ttl": 3,
      "attempts": [
        {"success": true, "resolvedAddress": "223.120.201.69"}
      ]
    }
  ]
}`)

	ips, err := parseNextTraceIPs(data)
	if err != nil {
		t.Fatalf("parseNextTraceIPs error = %v", err)
	}
	if len(ips) != 3 {
		t.Fatalf("len(ips) = %d, want 3", len(ips))
	}
	if got := ips[0].String(); got != "192.0.2.1" {
		t.Fatalf("first hop = %s", got)
	}
	if got := ips[1].String(); got != "218.105.2.89" {
		t.Fatalf("second hop = %s", got)
	}
	if got := ips[2].String(); got != "223.120.201.69" {
		t.Fatalf("third hop = %s", got)
	}
}

func TestParseNextTraceIPsRejectsMissingHops(t *testing.T) {
	if _, err := parseNextTraceIPs([]byte(`{"x":[]}`)); err == nil {
		t.Fatal("parseNextTraceIPs missing Hops error = nil")
	}
}
