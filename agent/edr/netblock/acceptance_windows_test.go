//go:build windows && edr_acceptance

package netblock

import (
	"net"
	"net/netip"
	"os"
	"testing"
	"time"
)

// The only traffic target and filter is loopback inside the disposable VM.
// The dynamic WFP session is always closed, including after test failures.
func TestWindowsAcceptanceWFPDynamicLoopback(t *testing.T) {
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("disposable Windows VM only")
	}
	listener, err := net.Listen("tcp4", "127.0.0.2:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for {
			c, e := listener.Accept()
			if e != nil {
				return
			}
			_ = c.Close()
		}
	}()
	dial := func() error {
		c, e := net.DialTimeout("tcp4", listener.Addr().String(), 750*time.Millisecond)
		if c != nil {
			c.Close()
		}
		return e
	}
	if err := dial(); err != nil {
		t.Fatalf("baseline loopback unreachable: %v", err)
	}
	blocker, err := NewOSBlocker("UTMStack disposable acceptance")
	if err != nil {
		t.Fatal(err)
	}
	b := blocker.(*wfpBlocker)
	defer procEngineClose.Call(uintptr(b.engine))
	addr := netip.MustParseAddr("127.0.0.2")
	if err := b.AddIP(addr, "out"); err != nil {
		t.Fatalf("real WFP AddIP: %v", err)
	}
	if len(b.filters[addr].ids) != 1 {
		t.Fatalf("native filter not retained: %+v", b.filters[addr])
	}
	blocked := dial() != nil
	if err := b.RemoveIP(addr); err != nil {
		t.Fatal(err)
	}
	if err := dial(); err != nil {
		t.Fatalf("loopback not restored after removal: %v", err)
	}
	prefix := netip.MustParsePrefix("127.0.0.2/32")
	if err := b.AddPrefix(prefix, "both"); err != nil {
		t.Fatalf("real WFP AddPrefix: %v", err)
	}
	if len(b.prefixFilters[prefix].ids) != 2 {
		t.Fatal("both-direction prefix did not install two filters")
	}
	if err := b.RemovePrefix(prefix); err != nil {
		t.Fatal(err)
	}
	if !blocked {
		t.Fatal("native WFP API accepted filters but outbound loopback traffic was not blocked; enforcement is not validated")
	}
	t.Log("native WFP blocked own loopback fixture and removal restored it")
}
