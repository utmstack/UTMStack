// edr/netblock/enforcer_test.go
package netblock

import (
	"net/netip"
	"sort"
	"testing"
)

type fakeBlocker struct {
	set map[netip.Addr]string
}

func newFake() *fakeBlocker { return &fakeBlocker{set: map[netip.Addr]string{}} }
func (f *fakeBlocker) AddIP(a netip.Addr, dir string) error { f.set[a] = dir; return nil }
func (f *fakeBlocker) RemoveIP(a netip.Addr) error          { delete(f.set, a); return nil }
func (f *fakeBlocker) Reset() error                         { f.set = map[netip.Addr]string{}; return nil }
func (f *fakeBlocker) Count() int                           { return len(f.set) }
func (f *fakeBlocker) keys() []string {
	var k []string
	for a := range f.set {
		k = append(k, a.String())
	}
	sort.Strings(k)
	return k
}

func TestEnforcerDiff(t *testing.T) {
	fb := newFake()
	e := NewEnforcer(fb, "both")
	a := netip.MustParseAddr("1.1.1.1")
	b := netip.MustParseAddr("2.2.2.2")
	cc := netip.MustParseAddr("3.3.3.3")

	if err := e.Apply([]netip.Addr{a, b}); err != nil {
		t.Fatal(err)
	}
	if fb.Count() != 2 {
		t.Fatalf("after first apply want 2, got %d", fb.Count())
	}
	// Drop b, add c: exactly one remove + one add, a untouched.
	if err := e.Apply([]netip.Addr{a, cc}); err != nil {
		t.Fatal(err)
	}
	got := fb.keys()
	if len(got) != 2 || got[0] != "1.1.1.1" || got[1] != "3.3.3.3" {
		t.Fatalf("after diff apply got %v", got)
	}
}
