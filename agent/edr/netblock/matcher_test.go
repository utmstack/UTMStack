// edr/netblock/matcher_test.go
package netblock

import (
	"net/netip"
	"testing"
)

func TestMatcherIPAndCIDR(t *testing.T) {
	m := NewMatcher()
	for _, s := range []struct {
		v, ty string
	}{{"1.2.3.4", "ip"}, {"10.0.0.0/8", "cidr"}, {"2001:db8::/32", "cidr"}} {
		ind, err := ParseIndicator(s.v, s.ty, 1)
		if err != nil {
			t.Fatalf("parse %s: %v", s.v, err)
		}
		m.Add(ind)
	}
	cases := []struct {
		ip   string
		want bool
	}{
		{"1.2.3.4", true},
		{"1.2.3.5", false},
		{"10.9.9.9", true},   // inside 10/8
		{"11.0.0.1", false},  // outside
		{"2001:db8::dead", true},
		{"2001:dc8::1", false},
	}
	for _, c := range cases {
		addr := netip.MustParseAddr(c.ip)
		_, ok := m.MatchIP(addr)
		if ok != c.want {
			t.Errorf("MatchIP(%s) = %v, want %v", c.ip, ok, c.want)
		}
	}
}

func TestMatcher4in6(t *testing.T) {
	m := NewMatcher()
	for _, s := range []struct {
		v, ty string
	}{{"1.2.3.4", "ip"}, {"10.0.0.0/8", "cidr"}} {
		ind, err := ParseIndicator(s.v, s.ty, 1)
		if err != nil {
			t.Fatalf("parse %s: %v", s.v, err)
		}
		m.Add(ind)
	}
	// A 4-in-6 form of the exact IPv4 indicator must still match.
	if _, ok := m.MatchIP(netip.MustParseAddr("::ffff:1.2.3.4")); !ok {
		t.Errorf("MatchIP(::ffff:1.2.3.4) = false, want true (4in6 exact IP)")
	}
	// A 4-in-6 form of an address inside the IPv4 CIDR must still match.
	if _, ok := m.MatchIP(netip.MustParseAddr("::ffff:10.9.9.9")); !ok {
		t.Errorf("MatchIP(::ffff:10.9.9.9) = false, want true (4in6 in IPv4 CIDR)")
	}
}

func TestParseIndicatorRejectsGarbage(t *testing.T) {
	if _, err := ParseIndicator("not-an-ip", "ip", 1); err == nil {
		t.Fatal("expected error for bad ip")
	}
}
