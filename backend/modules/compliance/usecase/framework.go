package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
)

type frameworkUsecase struct {
	controls   *ControlStore
	frameworks *FrameworkStore
}

func NewFrameworkUsecase(controls *ControlStore, frameworks *FrameworkStore) connectors.FrameworkUsecase {
	return &frameworkUsecase{controls: controls, frameworks: frameworks}
}

func (u *frameworkUsecase) ListControls(ctx context.Context) []domain.Control {
	return u.controls.All()
}

func (u *frameworkUsecase) GetControl(ctx context.Context, id string) (*domain.Control, error) {
	ctrl, ok := u.controls.Get(id)
	if !ok {
		return nil, domain.ErrControlNotFound
	}
	return ctrl, nil
}

func (u *frameworkUsecase) ListFrameworks(ctx context.Context) []domain.Framework {
	return u.frameworks.All()
}

func (u *frameworkUsecase) GetFramework(ctx context.Context, key string) (*domain.Framework, error) {
	fw, ok := u.frameworks.Get(key)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
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
	return u.controls.SetEnabled(id, enabled)
}

// ── Framework user-overlay CRUD ───────────────────────────────────────────────

func (u *frameworkUsecase) CreateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	return u.frameworks.Create(f)
}

func (u *frameworkUsecase) UpdateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	return u.frameworks.Update(f)
}

func (u *frameworkUsecase) DeleteFramework(ctx context.Context, key string) error {
	return u.frameworks.Delete(key)
}

func (u *frameworkUsecase) SetFrameworkEnabled(ctx context.Context, key string, enabled bool) error {
	return u.frameworks.SetEnabled(key, enabled)
}
