package netblock

import (
	"net/netip"
	"testing"
)

// A representation change (Evil.com -> evil.com) arrives as del(old)+add(new)
// that canonicalize to the same key. Dels-before-adds must keep it blocked.
func TestApplyDeltaSameKeyDelAddKeepsIndicator(t *testing.T) {
	evil, err := ParseIndicator("evil.com", "domain", 1)
	if err != nil {
		t.Fatal(err)
	}
	f := &Feed{store: NewStore()}
	f.cur = []Indicator{evil}

	// Order deliberately adversarial: add emitted before del (server order).
	f.applyDelta([]DailyOp{
		{Value: "evil.com", Type: "domain", Op: "add"},
		{Value: "Evil.com", Type: "domain", Op: "del"},
	})

	if _, ok := f.store.MatchName("evil.com"); !ok {
		t.Fatal("evil.com must remain blocked after a same-key del+add (dels before adds)")
	}
	if len(f.cur) != 1 {
		t.Fatalf("cur = %v, want the single canonical indicator", f.cur)
	}
}

// A genuine removal (del with no matching add) must still remove.
func TestApplyDeltaGenuineDelRemoves(t *testing.T) {
	ind, _ := ParseIndicator("9.9.9.9", "ip", 1)
	f := &Feed{store: NewStore()}
	f.cur = []Indicator{ind}
	f.applyDelta([]DailyOp{{Value: "9.9.9.9", Type: "ip", Op: "del"}})
	if _, ok := f.store.Match(netip.MustParseAddr("9.9.9.9")); ok {
		t.Fatal("9.9.9.9 must be removed by a genuine del")
	}
}
