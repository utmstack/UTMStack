// edr/netblock/allowlist_test.go
package netblock

import (
	"net/netip"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestAllowlistDefaults(t *testing.T) {
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = true
	al := BuildAllowlist(cfg, SystemNets{
		Resolvers: []netip.Addr{netip.MustParseAddr("8.8.8.8")},
		Gateways:  []netip.Addr{netip.MustParseAddr("192.168.1.1")},
	})
	must := map[string]bool{
		"127.0.0.1":   true, // loopback
		"::1":         true,
		"10.1.2.3":    true, // RFC1918 (private ranges on)
		"192.168.1.1": true, // gateway
		"8.8.8.8":     true, // resolver
		"169.254.1.1": true, // link-local
		"9.9.9.9":     false,
	}
	for ip, want := range must {
		if al.Allowed(netip.MustParseAddr(ip)) != want {
			t.Errorf("Allowed(%s) = %v, want %v", ip, !want, want)
		}
	}
}

func TestAllowlistOverlapsAllowed(t *testing.T) {
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = true
	al := BuildAllowlist(cfg, SystemNets{
		Resolvers: []netip.Addr{netip.MustParseAddr("8.8.8.8")},
	})
	cases := map[string]bool{
		"10.0.0.0/8": true,  // overlaps the 10/8 private range
		"8.8.8.0/24": true,  // contains the 8.8.8.8 resolver (exact IP)
		"9.9.9.0/24": false, // no overlap with any allowlisted IP/prefix
	}
	for cidr, want := range cases {
		p := netip.MustParsePrefix(cidr)
		if got := al.OverlapsAllowed(p); got != want {
			t.Errorf("OverlapsAllowed(%s) = %v, want %v", cidr, got, want)
		}
	}
}

func TestAllowlistPrivateOff(t *testing.T) {
	cfg := config.Default()
	cfg.Blocklist.AllowPrivateRanges = false
	al := BuildAllowlist(cfg, SystemNets{})
	if al.Allowed(netip.MustParseAddr("10.1.2.3")) {
		t.Fatal("10.1.2.3 must NOT be allowed when private ranges off")
	}
	if !al.Allowed(netip.MustParseAddr("127.0.0.1")) {
		t.Fatal("loopback always allowed")
	}
}
