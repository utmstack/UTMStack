// edr/netblock/store.go
package netblock

import (
	"net/netip"
	"sync/atomic"
)

// Store holds the active blocklist as an immutable *Matcher swapped atomically on
// each feed refresh, so readers never lock and never see a torn set.
type Store struct {
	cur atomic.Pointer[matcherState]
}

type matcherState struct {
	matcher   *Matcher
	published bool
}

func NewStore() *Store {
	s := &Store{}
	s.cur.Store(&matcherState{matcher: NewMatcher()})
	return s
}

func (s *Store) Swap(inds []Indicator) {
	m := NewMatcher()
	for _, ind := range inds {
		m.Add(ind)
	}
	s.cur.Store(&matcherState{matcher: m, published: true})
}

func (s *Store) Snapshot() *Matcher { return s.cur.Load().matcher }
func (s *Store) snapshotPublished() (*Matcher, bool) {
	state := s.cur.Load()
	return state.matcher, state.published
}

func (s *Store) Match(addr netip.Addr) (Indicator, bool) {
	return s.Snapshot().MatchIP(addr)
}

func (s *Store) MatchName(query string) (Indicator, bool) { return s.Snapshot().MatchName(query) }

func (s *Store) Count() int { return s.Snapshot().Len() }
