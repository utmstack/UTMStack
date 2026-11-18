// edr/netblock/reactive.go
package netblock

import (
	"errors"
	"net/netip"
	"sync"
	"time"
)

const MaxReactiveIPs = 16384

var ErrReactiveBudget = errors.New("reactive network filter budget exceeded")

type reactiveEntry struct {
	expires     time.Time
	allowRemove bool
	removeErr   error
}

// Reactive holds short-TTL WFP blocks for IPs resolved from blacklisted domains.
// Each Block (re)sets an addr's expiry to now+ttl; Sweep removes expired addrs.
// It owns only the addrs it added (distinct from the feed enforcer's set).
type Reactive struct {
	mu     sync.Mutex
	b      Blocker
	dir    string
	ttl    time.Duration
	now    func() time.Time
	expiry map[netip.Addr]reactiveEntry
	limit  int
	addErr error
	full   bool
	// onFeed reports whether an addr is now owned by the feed enforcer. When set,
	// Sweep hands off (instead of removing) an expired addr the feed has adopted,
	// since the enforcer shares the same blocker/filter and would otherwise lose it.
	onFeed func(netip.Addr) bool
}

func NewReactive(b Blocker, dir string, ttl time.Duration, now func() time.Time) *Reactive {
	return &Reactive{b: b, dir: dir, ttl: ttl, now: now, expiry: map[netip.Addr]reactiveEntry{}, limit: MaxReactiveIPs}
}

// SetOnFeed installs the predicate Sweep uses to detect feed-owned addrs.
func (r *Reactive) SetOnFeed(f func(netip.Addr) bool) {
	r.mu.Lock()
	r.onFeed = f
	r.mu.Unlock()
}

func (r *Reactive) Block(addrs []netip.Addr) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	exp := r.now().Add(r.ttl)
	active := 0
	r.addErr = nil
	for _, a := range addrs {
		a = a.Unmap()
		if !a.IsValid() {
			r.addErr = errors.New("invalid reactive IP address")
			continue
		}
		if _, ok := r.expiry[a]; !ok && len(r.expiry) >= r.limit {
			r.full = true
			continue
		}
		// AddIP is idempotent. Recheck the shared blocker even for an existing
		// lease, rather than claiming protection from an expiry record alone.
		if err := r.b.AddIP(a, r.dir); err != nil {
			r.addErr = err
			entry := r.expiry[a]
			entry.removeErr = err
			r.expiry[a] = entry // retain possible partial OS state for cleanup
			continue
		}
		r.expiry[a] = reactiveEntry{expires: exp}
		active++
	}
	return active
}

func (r *Reactive) Sweep() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := r.now()
	var firstErr error
	for a, entry := range r.expiry {
		if entry.allowRemove || now.After(entry.expires) {
			// If the feed enforcer now owns this addr, it holds the SAME shared
			// filter — hand off (drop our expiry) rather than removing the filter,
			// which the enforcer would not re-add (it's still in its applied set).
			if !entry.allowRemove && r.onFeed != nil && r.onFeed(a) {
				delete(r.expiry, a)
				continue
			}
			if err := r.b.RemoveIP(a); err != nil {
				entry.removeErr = err
				r.expiry[a] = entry
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			delete(r.expiry, a)
		}
	}
	if len(r.expiry) < r.limit {
		r.full = false
	}
	if len(r.expiry) == 0 {
		r.addErr = nil
	}
	return firstErr
}

// RemoveAllowed immediately removes newly exempted addresses. Failed removals
// remain tracked and are retried by later reconciliations/sweeps. The returned
// addresses were actually removed and can be forgotten by a shared owner.
func (r *Reactive) RemoveAllowed(allowed func(netip.Addr) bool) ([]netip.Addr, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var removed []netip.Addr
	var firstErr error
	for a, entry := range r.expiry {
		if !allowed(a) {
			if entry.allowRemove {
				entry.allowRemove, entry.removeErr = false, nil
				r.expiry[a] = entry
			}
			continue
		}
		if err := r.b.RemoveIP(a); err != nil {
			entry.allowRemove, entry.removeErr = true, err
			r.expiry[a] = entry
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		delete(r.expiry, a)
		removed = append(removed, a)
	}
	if len(r.expiry) < r.limit {
		r.full = false
	}
	if len(r.expiry) == 0 {
		r.addErr = nil
	}
	return removed, firstErr
}

// Owns reports an unexpired lease that should retain a shared exact-IP filter
// when another owner relinquishes it. Manager serializes ownership changes.
func (r *Reactive) Owns(a netip.Addr) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	entry, ok := r.expiry[a.Unmap()]
	return ok && !entry.allowRemove && entry.removeErr == nil && !r.now().After(entry.expires)
}

func (r *Reactive) HealthError() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, entry := range r.expiry {
		if entry.removeErr != nil {
			return entry.removeErr
		}
	}
	if r.full {
		return ErrReactiveBudget
	}
	return r.addErr
}

// Reset drops all reactive expiry tracking without touching the blocker (the
// caller has already cleared the shared filters, e.g. via Enforcer.Clear).
func (r *Reactive) Reset() {
	r.mu.Lock()
	r.expiry = map[netip.Addr]reactiveEntry{}
	r.addErr, r.full = nil, false
	r.mu.Unlock()
}

func (r *Reactive) Active() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.expiry)
}
