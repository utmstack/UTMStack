// edr/netblock/enforcer.go
package netblock

import (
	"errors"
	"net/netip"
)

// Blocker is the platform enforcement surface (WFP or Linux nftables).
type Blocker interface {
	AddIP(addr netip.Addr, dir string) error
	RemoveIP(addr netip.Addr) error
	AddPrefix(p netip.Prefix, dir string) error
	RemovePrefix(p netip.Prefix) error
	Reset() error
	Count() int
}

// ApplyPolicy uses a backend's atomic batch primitive when available. This keeps
// large Linux feeds to one bounded nft transaction and preserves reactive exact
// filters when ownership is handed off. Other platforms retain their existing
// per-filter reconciliation and failure behavior.
func (e *Enforcer) ApplyPolicy(desired []netip.Addr, prefixes []netip.Prefix) error {
	batch, ok := e.b.(interface {
		ApplyDelta([]netip.Addr, []netip.Addr, []netip.Prefix, []netip.Prefix, string) error
	})
	if !ok {
		return errors.Join(e.Apply(desired), e.ApplyPrefixes(prefixes))
	}
	want := make(map[netip.Addr]struct{}, len(desired))
	wantP := make(map[netip.Prefix]struct{}, len(prefixes))
	var add, remove []netip.Addr
	var addP, removeP []netip.Prefix
	for _, a := range desired {
		a = a.Unmap()
		if _, dup := want[a]; dup {
			continue
		}
		want[a] = struct{}{}
		if _, exists := e.applied[a]; !exists || e.uncertain[a] {
			add = append(add, a)
		}
	}
	for a := range e.applied {
		if _, exists := want[a]; !exists && (e.retainIP == nil || !e.retainIP(a)) {
			remove = append(remove, a)
		}
	}
	for _, p := range prefixes {
		p = p.Masked()
		if _, dup := wantP[p]; dup {
			continue
		}
		wantP[p] = struct{}{}
		if _, exists := e.appliedP[p]; !exists || e.uncertainP[p] {
			addP = append(addP, p)
		}
	}
	for p := range e.appliedP {
		if _, exists := wantP[p]; !exists {
			removeP = append(removeP, p)
		}
	}
	if len(add)+len(remove)+len(addP)+len(removeP) > 0 {
		if err := batch.ApplyDelta(add, remove, addP, removeP, e.dir); err != nil {
			for a := range e.applied {
				e.uncertain[a] = true
			}
			for p := range e.appliedP {
				e.uncertainP[p] = true
			}
			return err
		}
	}
	e.applied = want
	e.appliedP = wantP
	e.uncertain = map[netip.Addr]bool{}
	e.uncertainP = map[netip.Prefix]bool{}
	return nil
}

// Enforcer reconciles a desired IP set against what is currently applied, issuing
// the minimal add/remove calls to the Blocker.
type Enforcer struct {
	b          Blocker
	dir        string
	applied    map[netip.Addr]struct{}
	appliedP   map[netip.Prefix]struct{}
	uncertain  map[netip.Addr]bool
	uncertainP map[netip.Prefix]bool
	// retainIP lets the shared reactive owner keep a leased exact-IP filter.
	// The manager serializes this callback with all ownership changes.
	retainIP func(netip.Addr) bool
}

func NewEnforcer(b Blocker, dir string) *Enforcer {
	return &Enforcer{
		b:         b,
		dir:       dir,
		applied:   map[netip.Addr]struct{}{},
		appliedP:  map[netip.Prefix]struct{}{},
		uncertain: map[netip.Addr]bool{}, uncertainP: map[netip.Prefix]bool{},
	}
}

func (e *Enforcer) Apply(desired []netip.Addr) error {
	want := make(map[netip.Addr]struct{}, len(desired))
	for _, a := range desired {
		a = a.Unmap()
		want[a] = struct{}{}
		if _, ok := e.applied[a]; !ok || e.uncertain[a] {
			if err := e.b.AddIP(a, e.dir); err != nil {
				return err
			}
			e.applied[a] = struct{}{}
			delete(e.uncertain, a)
		}
	}
	for a := range e.applied {
		if _, ok := want[a]; !ok {
			if e.retainIP != nil && e.retainIP(a) {
				delete(e.applied, a)
				delete(e.uncertain, a)
				continue
			}
			if err := e.b.RemoveIP(a); err != nil {
				e.uncertain[a] = true
				return err
			}
			delete(e.applied, a)
			delete(e.uncertain, a)
		}
	}
	return nil
}

// OwnsIP is exact-filter ownership; a covering prefix is a distinct filter and
// cannot inherit a reactive exact-IP lease. Callers serialize with Apply/Clear.
func (e *Enforcer) OwnsIP(a netip.Addr) bool {
	_, ok := e.applied[a.Unmap()]
	return ok && !e.uncertain[a.Unmap()]
}

func (e *Enforcer) BlocksIP(a netip.Addr) bool {
	a = a.Unmap()
	if e.OwnsIP(a) {
		return true
	}
	for p := range e.appliedP {
		if !e.uncertainP[p] && p.Contains(a) {
			return true
		}
	}
	return false
}

// ApplyPrefixes reconciles a desired CIDR set against what is currently applied,
// issuing the minimal AddPrefix/RemovePrefix calls to the Blocker (mirrors Apply).
func (e *Enforcer) ApplyPrefixes(desired []netip.Prefix) error {
	want := make(map[netip.Prefix]struct{}, len(desired))
	for _, p := range desired {
		p = p.Masked()
		want[p] = struct{}{}
		if _, ok := e.appliedP[p]; !ok || e.uncertainP[p] {
			if err := e.b.AddPrefix(p, e.dir); err != nil {
				return err
			}
			e.appliedP[p] = struct{}{}
			delete(e.uncertainP, p)
		}
	}
	for p := range e.appliedP {
		if _, ok := want[p]; !ok {
			if err := e.b.RemovePrefix(p); err != nil {
				e.uncertainP[p] = true
				return err
			}
			delete(e.appliedP, p)
			delete(e.uncertainP, p)
		}
	}
	return nil
}

func (e *Enforcer) Clear() error {
	if err := e.b.Reset(); err != nil {
		for a := range e.applied {
			e.uncertain[a] = true
		}
		for p := range e.appliedP {
			e.uncertainP[p] = true
		}
		return err
	}
	e.applied = map[netip.Addr]struct{}{}
	e.appliedP = map[netip.Prefix]struct{}{}
	e.uncertain = map[netip.Addr]bool{}
	e.uncertainP = map[netip.Prefix]bool{}
	return nil
}

// BlockerCount reports the number of filters the underlying Blocker actually holds
// (0 for the no-op blocker used when WFP init failed), so status reflects real
// enforcement rather than the intended count.
func (e *Enforcer) BlockerCount() int { return e.b.Count() }
