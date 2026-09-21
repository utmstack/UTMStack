// edr/netblock/store.go
package netblock

import (
	"net/netip"
	"sync/atomic"
)

// Store holds the active blocklist as an immutable *Matcher swapped atomically on
// each feed refresh, so readers never lock and never see a torn set.
type Store struct {
	cur atomic.Pointer[Matcher]
}

func NewStore() *Store {
	s := &Store{}
	s.cur.Store(NewMatcher())
	return s
}

func (s *Store) Swap(inds []Indicator) {
	m := NewMatcher()
	for _, ind := range inds {
		m.Add(ind)
	}
	s.cur.Store(m)
}

func (s *Store) Snapshot() *Matcher { return s.cur.Load() }

func (s *Store) Match(addr netip.Addr) (Indicator, bool) {
	return s.cur.Load().MatchIP(addr)
}

func (s *Store) MatchName(query string) (Indicator, bool) { return s.cur.Load().MatchName(query) }

func (s *Store) Count() int { return s.cur.Load().Len() }
