package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type frameworkUsecase struct {
	controls   *ControlStore
	frameworks *FrameworkStore
	ent        *Entitlement
}

func NewFrameworkUsecase(controls *ControlStore, frameworks *FrameworkStore, ent *Entitlement) connectors.FrameworkUsecase {
	return &frameworkUsecase{controls: controls, frameworks: frameworks, ent: ent}
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
	for i := range out {
		out[i].Locked = u.ent.FrameworkLocked(&out[i])
	}
	return out
}

func (u *frameworkUsecase) GetFramework(ctx context.Context, key string) (*domain.Framework, error) {
	fw, ok := u.frameworks.Get(key)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	fw.Locked = u.ent.FrameworkLocked(fw)
	return fw, nil
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
	if fw, ok := u.frameworks.Get(key); ok && u.ent.FrameworkLocked(fw) {
		return domain.ErrFrameworkLocked
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
