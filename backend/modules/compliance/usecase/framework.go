package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type frameworkUsecase struct {
	controls   *ControlStore
	frameworks *FrameworkStore
	possessed  connectors.TenantFrameworkRepository // nil → single-tenant / no possession layer
	ent        *Entitlement
}

func NewFrameworkUsecase(controls *ControlStore, frameworks *FrameworkStore, possessed connectors.TenantFrameworkRepository, ent *Entitlement) connectors.FrameworkUsecase {
	return &frameworkUsecase{controls: controls, frameworks: frameworks, possessed: possessed, ent: ent}
}

func (u *frameworkUsecase) ListControls(ctx context.Context) []domain.Control {
	out := u.controls.All()
	for i := range out {
		out[i].Locked = u.ent.ControlLocked(&out[i])
	}
	return out
}

func (u *frameworkUsecase) GetControl(ctx context.Context, id string) (*domain.Control, error) {
	ctrl, ok := u.controls.Get(id)
	if !ok {
		return nil, domain.ErrControlNotFound
	}
	ctrl.Locked = u.ent.ControlLocked(ctrl)
	return ctrl, nil
}

func (u *frameworkUsecase) ListFrameworks(ctx context.Context) []domain.Framework {
	out := u.frameworks.All()
	// A tenant sees Enabled ⇔ they possess the framework (row present in
	// compliance_tenant_framework) AND the file itself isn't `.disabled` at
	// the platform level. Empty ctx tenant (on-prem/global) keeps the file
	// state — same behaviour the module had before possession existed.
	possessed := u.possessedSet(ctx)
	for i := range out {
		out[i].Locked = u.ent.FrameworkLocked(&out[i])
		if possessed != nil {
			out[i].Enabled = out[i].Enabled && possessed[out[i].Key]
		}
	}
	return out
}

func (u *frameworkUsecase) GetFramework(ctx context.Context, key string) (*domain.Framework, error) {
	fw, ok := u.frameworks.Get(key)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	fw.Locked = u.ent.FrameworkLocked(fw)
	if possessed := u.possessedSet(ctx); possessed != nil {
		fw.Enabled = fw.Enabled && possessed[fw.Key]
	}
	return fw, nil
}

// possessedSet returns the acting tenant's framework-key set, or nil for a
// global/on-prem caller (empty tenant) — nil means "don't overlay, take the
// file state as-is".
func (u *frameworkUsecase) possessedSet(ctx context.Context) map[string]bool {
	if u.possessed == nil || authz.TenantIDFromContext(ctx) == "" {
		return nil
	}
	keys, err := u.possessed.List(ctx)
	if err != nil {
		return map[string]bool{}
	}
	set := make(map[string]bool, len(keys))
	for _, k := range keys {
		set[k] = true
	}
	return set
}

// ── Control user-overlay CRUD ─────────────────────────────────────────────────

func (u *frameworkUsecase) CreateControl(ctx context.Context, c domain.Control) (*domain.Control, error) {
	return u.controls.Create(c)
}

func (u *frameworkUsecase) UpdateControl(ctx context.Context, c domain.Control) (*domain.Control, error) {
	return u.controls.Update(c)
}

func (u *frameworkUsecase) DeleteControl(ctx context.Context, id string) error {
	return u.controls.Delete(id)
}

func (u *frameworkUsecase) SetControlEnabled(ctx context.Context, id string, enabled bool) error {
	if ctrl, ok := u.controls.Get(id); ok && u.ent.ControlLocked(ctrl) {
		return domain.ErrControlLocked
	}
	return u.controls.SetEnabled(id, enabled)
}

// ── Framework user-overlay CRUD ───────────────────────────────────────────────

func (u *frameworkUsecase) CreateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if err := u.assertReferencedControlsAllowed(f); err != nil {
		return nil, err
	}
	return u.frameworks.Create(f)
}

func (u *frameworkUsecase) UpdateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if err := u.assertReferencedControlsAllowed(f); err != nil {
		return nil, err
	}
	return u.frameworks.Update(f)
}

func (u *frameworkUsecase) DeleteFramework(ctx context.Context, key string) error {
	return u.frameworks.Delete(key)
}

func (u *frameworkUsecase) SetFrameworkEnabled(ctx context.Context, key string, enabled bool) error {
	fw, ok := u.frameworks.Get(key)
	if !ok {
		return domain.ErrFrameworkNotFound
	}
	if u.ent.FrameworkLocked(fw) {
		return domain.ErrFrameworkLocked
	}
	// Multi-tenant: toggle the possession row for the acting tenant. The file's
	// .disabled state stays as the platform-wide off switch and is only touched
	// by an on-prem/global caller (empty ctx tenant), matching the pre-
	// possession behaviour so a single-tenant install keeps working.
	if u.possessed != nil && authz.TenantIDFromContext(ctx) != "" {
		if enabled {
			return u.possessed.Enable(ctx, key)
		}
		return u.possessed.Disable(ctx, key)
	}
	return u.frameworks.SetEnabled(key, enabled)
}

func (u *frameworkUsecase) assertReferencedControlsAllowed(f domain.Framework) error {
	for si := range f.Sections {
		for ri := range f.Sections[si].Requirements {
			for _, id := range f.Sections[si].Requirements[ri].SatisfiedBy {
				if u.ent.systemControlIDLocked(id, u.controls.Get) {
					return domain.ErrControlLocked
				}
			}
		}
	}
	return nil
}
