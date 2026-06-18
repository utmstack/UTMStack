package usecase

import (
	"sync"

	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type Entitlement struct {
	isEnterprise func() bool // fallback edition source (billing license)

	mu         sync.RWMutex
	set        bool // true once the entitlements plugin has pushed
	enterprise bool
	frameworks map[string]bool
	controls   map[string]bool
}

func NewEntitlement(isEnterprise func() bool) *Entitlement {
	return &Entitlement{isEnterprise: isEnterprise}
}

func (e *Entitlement) Set(enterprise bool, frameworks, controls []string) {
	fw := make(map[string]bool, len(frameworks))
	for _, k := range frameworks {
		fw[k] = true
	}
	ct := make(map[string]bool, len(controls))
	for _, id := range controls {
		ct[id] = true
	}
	e.mu.Lock()
	e.set, e.enterprise, e.frameworks, e.controls = true, enterprise, fw, ct
	e.mu.Unlock()
}

func (e *Entitlement) effectiveEnterprise() bool {
	if e == nil {
		return true
	}
	e.mu.RLock()
	set, ent := e.set, e.enterprise
	e.mu.RUnlock()
	if set {
		return ent
	}
	return e.isEnterprise == nil || e.isEnterprise()
}

func (e *Entitlement) frameworkAllowed(key string) bool {
	e.mu.RLock()
	set := e.set
	allowed := e.frameworks[key]
	e.mu.RUnlock()
	if set {
		return allowed
	}
	return communityFrameworks[key]
}

func (e *Entitlement) controlAllowed(id string) bool {
	e.mu.RLock()
	set := e.set
	allowed := e.controls[id]
	e.mu.RUnlock()
	if set {
		return allowed
	}
	return communityControls[id]
}

func (e *Entitlement) FrameworkLocked(f *domain.Framework) bool {
	if e.effectiveEnterprise() || f == nil || !f.System {
		return false
	}
	return !e.frameworkAllowed(f.Key)
}

func (e *Entitlement) ControlLocked(c *domain.Control) bool {
	if e.effectiveEnterprise() || c == nil || !c.System {
		return false
	}
	return !e.controlAllowed(c.ID)
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
