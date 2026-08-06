package usecase

import (
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

// Entitlement decides what a deployment without an enterprise licence may use.
// The licence is read live from billing, so a lapse or a renewal takes effect on
// its next refresh without anything having to be pushed in.
type Entitlement struct {
	isEnterprise func() bool
}

func NewEntitlement(isEnterprise func() bool) *Entitlement {
	return &Entitlement{isEnterprise: isEnterprise}
}

func (e *Entitlement) effectiveEnterprise() bool {
	if e == nil {
		return true
	}
	return e.isEnterprise == nil || e.isEnterprise()
}

func (e *Entitlement) FrameworkLocked(f *domain.Framework) bool {
	if e.effectiveEnterprise() || f == nil || !f.System {
		return false
	}
	return !communityFrameworks[f.Key]
}

func (e *Entitlement) ControlLocked(c *domain.Control) bool {
	if e.effectiveEnterprise() || c == nil || !c.System {
		return false
	}
	return !communityControls[c.ID]
}

func (e *Entitlement) systemControlIDLocked(id string, lookup func(string) (*domain.Control, bool)) bool {
	if e.effectiveEnterprise() {
		return false
	}
	c, ok := lookup(id)
	if !ok {
		return false
	}
	return e.ControlLocked(c)
}
