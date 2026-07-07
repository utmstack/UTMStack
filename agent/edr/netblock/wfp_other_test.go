//go:build !windows

package netblock

import (
	"net/netip"
	"testing"
)

func TestNoopBlocker(t *testing.T) {
	b, err := NewOSBlocker("UTMStackEDR-Blocklist")
	if err != nil {
		t.Fatal(err)
	}
	if err := b.AddIP(netip.MustParseAddr("1.2.3.4"), "both"); err != nil {
		t.Fatal(err)
	}
	if b.Count() != 0 { // no-op counts nothing
		t.Fatalf("noop count = %d", b.Count())
	}
	if err := b.Reset(); err != nil {
		t.Fatal(err)
	}
}
