package usecase

import (
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

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

func (e *Entitlement) FrameworkLocked(f *domain.Framework, system bool) bool {
	if e.effectiveEnterprise() || f == nil || !system {
		return false
	}
	return !communityFrameworks[f.Key]
}

func (e *Entitlement) ControlLocked(c *domain.Control, system bool) bool {
	if e.effectiveEnterprise() || c == nil || !system {
		return false
	}
	return !communityControls[c.ID]
}
