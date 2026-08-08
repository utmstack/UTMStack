package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/modules/compliance/dto"
)

type frameworkUsecase struct {
	controls   *ControlStore
	frameworks *FrameworkStore
	ent        *Entitlement
}

func NewFrameworkUsecase(controls *ControlStore, frameworks *FrameworkStore, ent *Entitlement) connectors.FrameworkUsecase {
	return &frameworkUsecase{controls: controls, frameworks: frameworks, ent: ent}
}

func (u *frameworkUsecase) ListControls(ctx context.Context) []dto.ControlResponse {
	all := u.controls.All(ctx)
	out := make([]dto.ControlResponse, 0, len(all))
	for i := range all {
		system := u.controls.IsSystem(all[i].ID)
		out = append(out, dto.ControlResponse{
			Control: all[i],
			System:  system,
			Locked:  u.ent.ControlLocked(&all[i], system),
		})
	}
	return out
}

func (u *frameworkUsecase) GetControl(ctx context.Context, id string) (*dto.ControlResponse, error) {
	c, ok := u.controls.Get(ctx, id)
	if !ok {
		return nil, domain.ErrControlNotFound
	}
	system := u.controls.IsSystem(id)
	return &dto.ControlResponse{Control: *c, System: system, Locked: u.ent.ControlLocked(c, system)}, nil
}

func (u *frameworkUsecase) ListFrameworks(ctx context.Context) []dto.FrameworkResponse {
	all := u.frameworks.All(ctx)
	out := make([]dto.FrameworkResponse, 0, len(all))
	for i := range all {
		system := u.frameworks.IsSystem(all[i].Key)
		out = append(out, dto.FrameworkResponse{
			Framework: all[i],
			System:    system,
			Locked:    u.ent.FrameworkLocked(&all[i], system),
		})
	}
	return out
}

func (u *frameworkUsecase) GetFramework(ctx context.Context, key string) (*dto.FrameworkResponse, error) {
	fw, ok := u.frameworks.Get(ctx, key)
	if !ok {
		return nil, domain.ErrFrameworkNotFound
	}
	system := u.frameworks.IsSystem(key)
	return &dto.FrameworkResponse{Framework: *fw, System: system, Locked: u.ent.FrameworkLocked(fw, system)}, nil
}

func (u *frameworkUsecase) CreateControl(ctx context.Context, c domain.Control) (*domain.Control, error) {
	return u.controls.Create(ctx, c)
}

func (u *frameworkUsecase) UpdateControl(ctx context.Context, c domain.Control) (*domain.Control, error) {
	return u.controls.Update(ctx, c)
}

func (u *frameworkUsecase) DeleteControl(ctx context.Context, id string) error {
	return u.controls.Delete(ctx, id)
}

func (u *frameworkUsecase) CreateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if err := u.assertReferencedControlsAllowed(ctx, f); err != nil {
		return nil, err
	}
	return u.frameworks.Create(ctx, f)
}

func (u *frameworkUsecase) UpdateFramework(ctx context.Context, f domain.Framework) (*domain.Framework, error) {
	if err := u.assertReferencedControlsAllowed(ctx, f); err != nil {
		return nil, err
	}
	return u.frameworks.Update(ctx, f)
}

func (u *frameworkUsecase) DeleteFramework(ctx context.Context, key string) error {
	return u.frameworks.Delete(ctx, key)
}

// assertReferencedControlsAllowed stops a community tenant from reaching a
// licensed control by naming it from a framework of their own.
func (u *frameworkUsecase) assertReferencedControlsAllowed(ctx context.Context, f domain.Framework) error {
	for si := range f.Sections {
		for ri := range f.Sections[si].Requirements {
			for _, id := range f.Sections[si].Requirements[ri].SatisfiedBy {
				c, ok := u.controls.Get(ctx, id)
				if !ok {
					continue
				}
				if u.ent.ControlLocked(c, u.controls.IsSystem(id)) {
					return domain.ErrControlLocked
				}
			}
		}
	}
	return nil
}
