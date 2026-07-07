// edr/netblock/enforcer.go
package netblock

import "net/netip"

// Blocker is the platform enforcement surface (WFP on Windows, no-op elsewhere).
type Blocker interface {
	AddIP(addr netip.Addr, dir string) error
	RemoveIP(addr netip.Addr) error
	Reset() error
	Count() int
}

// Enforcer reconciles a desired IP set against what is currently applied, issuing
// the minimal add/remove calls to the Blocker.
type Enforcer struct {
	b       Blocker
	dir     string
	applied map[netip.Addr]struct{}
}

func NewEnforcer(b Blocker, dir string) *Enforcer {
	return &Enforcer{b: b, dir: dir, applied: map[netip.Addr]struct{}{}}
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

func (e *Enforcer) Clear() error {
	if err := e.b.Reset(); err != nil {
		return err
	}
	e.applied = map[netip.Addr]struct{}{}
	return nil
}
