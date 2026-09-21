// edr/netblock/namematch_test.go
package netblock

import "testing"

func TestNameSetMatch(t *testing.T) {
	ns := NewNameSet()
	ns.Add("evil.com")
	ns.Add("bad.example.org")
	cases := []struct {
		q    string
		want bool
	}{
		{"evil.com", true},
		{"c2.evil.com", true},     // subdomain
		{"a.b.evil.com", true},    // deep subdomain
		{"EVIL.COM", true},        // case-insensitive
		{"evil.com.", true},       // trailing dot
		{"notevil.com", false},    // not a label-boundary match
		{"evil.com.attacker.net", false}, // evil.com is a label, but not a suffix-parent here
		{"example.org", false},    // parent of a listed host is not listed
		{"bad.example.org", true},
		{"x.bad.example.org", true},
		{"good.com", false},
	}
	for _, c := range cases {
		_, ok := ns.Match(c.q)
		if ok != c.want {
			t.Errorf("Match(%q) = %v, want %v", c.q, ok, c.want)
		}
	}
	if ns.Len() != 2 {
		t.Fatalf("Len = %d, want 2", ns.Len())
	}
}
