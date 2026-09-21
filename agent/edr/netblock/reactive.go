// edr/netblock/reactive.go
package netblock

import (
	"net/netip"
	"sync"
	"time"
)

// Reactive holds short-TTL WFP blocks for IPs resolved from blacklisted domains.
// Each Block (re)sets an addr's expiry to now+ttl; Sweep removes expired addrs.
// It owns only the addrs it added (distinct from the feed enforcer's set).
type Reactive struct {
	mu     sync.Mutex
	b      Blocker
	dir    string
	ttl    time.Duration
	now    func() time.Time
	expiry map[netip.Addr]time.Time
	// onFeed reports whether an addr is now owned by the feed enforcer. When set,
	// Sweep hands off (instead of removing) an expired addr the feed has adopted,
	// since the enforcer shares the same blocker/filter and would otherwise lose it.
	onFeed func(netip.Addr) bool
}

func NewReactive(b Blocker, dir string, ttl time.Duration, now func() time.Time) *Reactive {
	return &Reactive{b: b, dir: dir, ttl: ttl, now: now, expiry: map[netip.Addr]time.Time{}}
}

// SetOnFeed installs the predicate Sweep uses to detect feed-owned addrs.
func (r *Reactive) SetOnFeed(f func(netip.Addr) bool) {
	r.mu.Lock()
	r.onFeed = f
	r.mu.Unlock()
}

func (r *Reactive) Block(addrs []netip.Addr) {
	r.mu.Lock()
	defer r.mu.Unlock()
	exp := r.now().Add(r.ttl)
	for _, a := range addrs {
		if _, ok := r.expiry[a]; !ok {
			if err := r.b.AddIP(a, r.dir); err != nil {
				continue
			}
		}
		r.expiry[a] = exp
	}
}

func (r *Reactive) Sweep() {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	for a, exp := range r.expiry {
		if now.After(exp) {
			// If the feed enforcer now owns this addr, it holds the SAME shared
			// filter — hand off (drop our expiry) rather than removing the filter,
			// which the enforcer would not re-add (it's still in its applied set).
			if r.onFeed != nil && r.onFeed(a) {
				delete(r.expiry, a)
				continue
			}
			_ = r.b.RemoveIP(a)
			delete(r.expiry, a)
		}
	}
}

// Reset drops all reactive expiry tracking without touching the blocker (the
// caller has already cleared the shared filters, e.g. via Enforcer.Clear).
func (r *Reactive) Reset() {
	r.mu.Lock()
	r.expiry = map[netip.Addr]time.Time{}
	r.mu.Unlock()
}

func (r *Reactive) Active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.expiry)
}
