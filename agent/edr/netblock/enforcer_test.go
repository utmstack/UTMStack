// edr/netblock/enforcer_test.go
package netblock

import (
	"net/netip"
	"sort"
	"testing"
)

type fakeBlocker struct {
	set  map[netip.Addr]string
	pset map[netip.Prefix]string
}

func newFake() *fakeBlocker {
	return &fakeBlocker{set: map[netip.Addr]string{}, pset: map[netip.Prefix]string{}}
}

// newFakeP is the prefix-aware constructor; identical to newFake but named to
// signal that a test exercises the CIDR path.
func newFakeP() *fakeBlocker {
	return &fakeBlocker{set: map[netip.Addr]string{}, pset: map[netip.Prefix]string{}}
}

func (f *fakeBlocker) AddIP(a netip.Addr, dir string) error { f.set[a] = dir; return nil }
func (f *fakeBlocker) RemoveIP(a netip.Addr) error          { delete(f.set, a); return nil }
func (f *fakeBlocker) AddPrefix(p netip.Prefix, dir string) error {
	f.pset[p] = dir
	return nil
}
func (f *fakeBlocker) RemovePrefix(p netip.Prefix) error { delete(f.pset, p); return nil }
func (f *fakeBlocker) Reset() error {
	f.set = map[netip.Addr]string{}
	f.pset = map[netip.Prefix]string{}
	return nil
}
func (f *fakeBlocker) Count() int { return len(f.set) + len(f.pset) }
func (f *fakeBlocker) keys() []string {
	var k []string
	for a := range f.set {
		k = append(k, a.String())
	}
	sort.Strings(k)
	return k
}
func (f *fakeBlocker) pcount() int              { return len(f.pset) }
func (f *fakeBlocker) hasP(p netip.Prefix) bool { _, ok := f.pset[p]; return ok }
func (f *fakeBlocker) pkeys() []string {
	var k []string
	for p := range f.pset {
		k = append(k, p.String())
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

func TestEnforcerPrefixDiff(t *testing.T) {
	fb := newFakeP() // prefix-aware fake below
	e := NewEnforcer(fb, "both")
	a := netip.MustParsePrefix("10.0.0.0/8")
	b := netip.MustParsePrefix("192.168.0.0/16")
	cc := netip.MustParsePrefix("172.16.0.0/12")
	if err := e.ApplyPrefixes([]netip.Prefix{a, b}); err != nil {
		t.Fatal(err)
	}
	if fb.pcount() != 2 {
		t.Fatalf("want 2 prefixes, got %d", fb.pcount())
	}
	if err := e.ApplyPrefixes([]netip.Prefix{a, cc}); err != nil {
		t.Fatal(err)
	}
	if fb.pcount() != 2 || !fb.hasP(a) || !fb.hasP(cc) || fb.hasP(b) {
		t.Fatalf("diff wrong: %v", fb.pkeys())
	}
}
