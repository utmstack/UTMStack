// edr/netblock/reactive_test.go
package netblock

import (
	"net/netip"
	"testing"
	"time"
)

func TestReactiveTTL(t *testing.T) {
	fb := newFake() // from enforcer_test.go (same package)
	base := time.Unix(1000, 0)
	clk := base
	r := NewReactive(fb, "both", 30*time.Minute, func() time.Time { return clk })

	a := netip.MustParseAddr("1.2.3.4")
	r.Block([]netip.Addr{a})
	if fb.Count() != 1 || r.Active() != 1 {
		t.Fatalf("after Block: blocker=%d active=%d", fb.Count(), r.Active())
	}
	// Before TTL: sweep keeps it.
	clk = base.Add(20 * time.Minute)
	r.Sweep()
	if fb.Count() != 1 {
		t.Fatal("must survive before TTL")
	}
	// Re-block refreshes expiry.
	r.Block([]netip.Addr{a})
	clk = base.Add(40 * time.Minute) // 20m since refresh
	r.Sweep()
	if fb.Count() != 1 {
		t.Fatal("re-block should have refreshed TTL")
	}
	// Past TTL from last refresh: swept.
	clk = base.Add(75 * time.Minute)
	r.Sweep()
	if fb.Count() != 0 || r.Active() != 0 {
		t.Fatalf("must expire: blocker=%d active=%d", fb.Count(), r.Active())
	}
}

// TestReactiveHandoffToFeed verifies that when a reactive block expires but the
// feed enforcer has since adopted the addr (shared blocker/filter), Sweep hands
// off — dropping its expiry without removing the shared filter — and that when the
// feed does NOT own it, Sweep still removes the filter.
func TestReactiveHandoffToFeed(t *testing.T) {
	a := netip.MustParseAddr("9.9.9.9")

	// Case 1: feed now owns the addr → Sweep hands off (filter survives).
	{
		fb := newFake()
		base := time.Unix(2000, 0)
		clk := base
		r := NewReactive(fb, "both", 30*time.Minute, func() time.Time { return clk })
		r.SetOnFeed(func(x netip.Addr) bool { return x == a })
		r.Block([]netip.Addr{a})
		if fb.Count() != 1 || r.Active() != 1 {
			t.Fatalf("after Block: blocker=%d active=%d", fb.Count(), r.Active())
		}
		clk = base.Add(40 * time.Minute) // past TTL
		r.Sweep()
		if fb.Count() != 1 {
			t.Fatalf("feed-owned filter must survive handoff: blocker=%d", fb.Count())
		}
		if r.Active() != 0 {
			t.Fatalf("reactive must release its expiry on handoff: active=%d", r.Active())
		}
	}

	// Case 2: feed does NOT own the addr → Sweep removes the filter (baseline).
	{
		fb := newFake()
		base := time.Unix(2000, 0)
		clk := base
		r := NewReactive(fb, "both", 30*time.Minute, func() time.Time { return clk })
		r.SetOnFeed(func(netip.Addr) bool { return false })
		r.Block([]netip.Addr{a})
		clk = base.Add(40 * time.Minute) // past TTL
		r.Sweep()
		if fb.Count() != 0 || r.Active() != 0 {
			t.Fatalf("non-feed filter must be removed: blocker=%d active=%d", fb.Count(), r.Active())
		}
	}
}
