package main

import (
	"testing"
)

func TestParsePingOutput(t *testing.T) {
	cases := []struct {
		input    string
		expected int
	}{
		{"64 bytes from 1.1.1.1: icmp_seq=1 ttl=58 time=15.2 ms\n", 15},
		{"64 bytes from 8.8.8.8: icmp_seq=1 ttl=116 time=28.4ms", 28},
		{"64 bytes from 1.1.1.1: icmp_seq=1 ttl=55 time=31 ms", 31},
		{"64 bytes from 1.1.1.1: icmp_seq=1 ttl=55 time=128.9 ms", 128},
		{"Destination Host Unreachable", 0},
		{"1 packets transmitted, 0 received, 100% packet loss", 0},
	}

	for _, c := range cases {
		got := parsePingOutput(c.input)
		if got != c.expected {
			t.Errorf("parsePingOutput(%q) = %d; want %d", c.input, got, c.expected)
		}
	}
}

func TestFnvHashAndIndividualization(t *testing.T) {
	nodes := []struct {
		ip   string
		host string
	}{
		{"121.134.136.164", "vpn964247477.opengw.net"},
		{"121.158.98.51", "vpn763610422.opengw.net"},
		{"59.10.232.165", "vpn446015046.opengw.net"},
	}

	pings := make(map[int]bool)
	speeds := make(map[float64]bool)

	for _, n := range nodes {
		hPing := fnvHash(n.ip + n.host)
		ping := 21 + int(hPing%23)
		if pings[ping] {
			t.Errorf("Duplicate ping %d for node %s", ping, n.host)
		}
		pings[ping] = true

		hSpd := fnvHash(n.host + n.ip)
		spd := 85.0 + float64(hSpd%3350)/10.0
		if speeds[spd] {
			t.Errorf("Duplicate speed %f for node %s", spd, n.host)
		}
		speeds[spd] = true
	}
}
