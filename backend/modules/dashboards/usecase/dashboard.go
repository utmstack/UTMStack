package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"

	"github.com/google/uuid"
)

type dashboardUsecase struct {
	repo connectors.DashboardRepository
}

func NewDashboardUsecase(repo connectors.DashboardRepository) connectors.DashboardUsecase {
	return &dashboardUsecase{repo: repo}
}

func (u *dashboardUsecase) Create(ctx context.Context, d *domain.Dashboard, user string) (*domain.Dashboard, error) {
	if d.ID != uuid.Nil {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(d.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	now := time.Now().UTC()
	d.CreatedDate = now
	d.ModifiedDate = now
	if err := u.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (u *dashboardUsecase) Update(ctx context.Context, d *domain.Dashboard, user string) (*domain.Dashboard, error) {
	if d.ID == uuid.Nil {
		return nil, domain.ErrIDRequired
	}
	existing, err := u.repo.FindByID(ctx, d.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	if existing.SystemOwner {
		return nil, domain.ErrSystemOwned
	}
	// Preserve creation metadata; never let an update rewrite it.
	d.CreatedDate = existing.CreatedDate
	d.SystemOwner = existing.SystemOwner
	d.ModifiedDate = time.Now().UTC()
	if err := u.repo.Save(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (u *dashboardUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Dashboard, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *dashboardUsecase) List(ctx context.Context, f dto.DashboardFilter) ([]domain.Dashboard, int64, error) {
	return u.repo.List(ctx, f)
}

// Delete removes a tenant-owned dashboard outright. A system-owned one can't
// be hard-deleted — DashboardBootstrap reseeds it by name on every boot — so
// instead it's dismissed: hidden from List and skipped by the next reseed.
// There's no undo; a dismissed dashboard's row stays around only so it can
// still be picked as a copy source (List with Dismissed=true) when creating
// a new one.
func (u *dashboardUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}
	if existing.SystemOwner {
		now := time.Now().UTC()
		existing.DismissedAt = &now
		existing.ModifiedDate = now
		return u.repo.Save(ctx, existing)
	}
	return u.repo.Delete(ctx, id)
}
