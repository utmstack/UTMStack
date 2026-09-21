// edr/netblock/enforcer.go
package netblock

import "net/netip"

// Blocker is the platform enforcement surface (WFP on Windows, no-op elsewhere).
type Blocker interface {
	AddIP(addr netip.Addr, dir string) error
	RemoveIP(addr netip.Addr) error
	AddPrefix(p netip.Prefix, dir string) error
	RemovePrefix(p netip.Prefix) error
	Reset() error
	Count() int
}

// Enforcer reconciles a desired IP set against what is currently applied, issuing
// the minimal add/remove calls to the Blocker.
type Enforcer struct {
	b        Blocker
	dir      string
	applied  map[netip.Addr]struct{}
	appliedP map[netip.Prefix]struct{}
}

func NewEnforcer(b Blocker, dir string) *Enforcer {
	return &Enforcer{
		b:        b,
		dir:      dir,
		applied:  map[netip.Addr]struct{}{},
		appliedP: map[netip.Prefix]struct{}{},
	}
}

func (e *Enforcer) Apply(desired []netip.Addr) error {
	want := make(map[netip.Addr]struct{}, len(desired))
	for _, a := range desired {
		want[a] = struct{}{}
		if _, ok := e.applied[a]; !ok {
			if err := e.b.AddIP(a, e.dir); err != nil {
				return err
			}
			e.applied[a] = struct{}{}
		}
	}
	for a := range e.applied {
		if _, ok := want[a]; !ok {
			if err := e.b.RemoveIP(a); err != nil {
				return err
			}
			delete(e.applied, a)
		}
	}
	return nil
}

// ApplyPrefixes reconciles a desired CIDR set against what is currently applied,
// issuing the minimal AddPrefix/RemovePrefix calls to the Blocker (mirrors Apply).
func (e *Enforcer) ApplyPrefixes(desired []netip.Prefix) error {
	want := make(map[netip.Prefix]struct{}, len(desired))
	for _, p := range desired {
		p = p.Masked()
		want[p] = struct{}{}
		if _, ok := e.appliedP[p]; !ok {
			if err := e.b.AddPrefix(p, e.dir); err != nil {
				return err
			}
			e.appliedP[p] = struct{}{}
		}
	}
	for p := range e.appliedP {
		if _, ok := want[p]; !ok {
			if err := e.b.RemovePrefix(p); err != nil {
				return err
			}
			delete(e.appliedP, p)
		}
	}
	return nil
}

func (e *Enforcer) Clear() error {
	if err := e.b.Reset(); err != nil {
		return err
	}
	e.applied = map[netip.Addr]struct{}{}
	e.appliedP = map[netip.Prefix]struct{}{}
	return nil
}

// BlockerCount reports the number of filters the underlying Blocker actually holds
// (0 for the no-op blocker used when WFP init failed), so status reflects real
// enforcement rather than the intended count.
func (e *Enforcer) BlockerCount() int { return e.b.Count() }
