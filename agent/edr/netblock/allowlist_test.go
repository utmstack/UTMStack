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
