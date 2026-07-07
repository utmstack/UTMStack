// edr/netblock/matcher.go
package netblock

import "net/netip"

// Matcher answers IP membership against single IPs and CIDR prefixes. Not safe
// for concurrent mutation; build once, then read (the Store swaps whole Matchers).
type Matcher struct {
	exact    map[netip.Addr]Indicator
	prefixes []prefixEntry
}

type prefixEntry struct {
	p   netip.Prefix
	ind Indicator
}

func NewMatcher() *Matcher {
	return &Matcher{exact: map[netip.Addr]Indicator{}}
}

func (m *Matcher) Add(ind Indicator) {
	switch ind.Type {
	case "ip":
		if a, err := netip.ParseAddr(ind.Value); err == nil {
			m.exact[a.Unmap()] = ind
		}
	case "cidr":
		if p, err := netip.ParsePrefix(ind.Value); err == nil {
			m.prefixes = append(m.prefixes, prefixEntry{p: p.Masked(), ind: ind})
		}
	}
}

func (m *Matcher) Len() int { return len(m.exact) + len(m.prefixes) }

// MatchIP returns the matched indicator (exact wins over prefix) and true if the
// address is blacklisted.
func (m *Matcher) MatchIP(addr netip.Addr) (Indicator, bool) {
	addr = addr.Unmap()
	if ind, ok := m.exact[addr]; ok {
		return ind, true
	}
	for _, pe := range m.prefixes {
		if pe.p.Contains(addr) {
			return pe.ind, true
		}
	}
	return Indicator{}, false
}
