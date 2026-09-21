// edr/netblock/namematch.go
package netblock

import "strings"

// NameSet answers domain/hostname membership with label-boundary suffix matching:
// a query matches a listed name if it equals it or is a subdomain of it. Built
// once, then read (the Store swaps whole matchers).
type NameSet struct {
	names map[string]struct{}
}

func NewNameSet() *NameSet { return &NameSet{names: map[string]struct{}{}} }

func canonName(s string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(s), "."))
}

func (n *NameSet) Add(name string) {
	if c := canonName(name); c != "" {
		n.names[c] = struct{}{}
	}
}

func (n *NameSet) Len() int { return len(n.names) }

// Match returns the listed name that the query equals or is a subdomain of.
// Walks the query's parent domains: for "a.b.evil.com" it tests "a.b.evil.com",
// "b.evil.com", "evil.com", "com" — the first that is listed wins.
func (n *NameSet) Match(query string) (string, bool) {
	q := canonName(query)
	if q == "" {
		return "", false
	}
	for {
		if _, ok := n.names[q]; ok {
			return q, true
		}
		i := strings.IndexByte(q, '.')
		if i < 0 {
			return "", false
		}
		q = q[i+1:]
	}
}
