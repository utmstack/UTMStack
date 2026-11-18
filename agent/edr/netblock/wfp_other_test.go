//go:build !windows && !linux

package netblock

import (
	"errors"
	"net/netip"
	"testing"
)

func TestUnsupportedBlockerNeverReportsSuccess(t *testing.T) {
	b, err := NewOSBlocker("UTMStackEDR-Blocklist")
	if b != nil || !errors.Is(err, ErrUnsupported) {
		t.Fatalf("unsupported constructor: %v %v", b, err)
	}
	unavailable := unavailableBlocker{}
	for _, err := range []error{unavailable.AddIP(netip.MustParseAddr("192.0.2.1"), "both"), unavailable.RemoveIP(netip.MustParseAddr("192.0.2.1")), unavailable.AddPrefix(netip.MustParsePrefix("192.0.2.0/24"), "both"), unavailable.RemovePrefix(netip.MustParsePrefix("192.0.2.0/24")), unavailable.Reset()} {
		if !errors.Is(err, ErrUnsupported) {
			t.Fatalf("false success: %v", err)
		}
	}
}
