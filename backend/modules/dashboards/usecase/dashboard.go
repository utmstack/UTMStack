package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type dashboardUsecase struct {
	repo connectors.DashboardRepository
}

func NewDashboardUsecase(repo connectors.DashboardRepository) connectors.DashboardUsecase {
	return &dashboardUsecase{repo: repo}
}

func (u *dashboardUsecase) Create(ctx context.Context, d *domain.Dashboard, user string) (*domain.Dashboard, error) {
	if d.ID != 0 {
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
	if d.ID == 0 {
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

func (u *dashboardUsecase) GetByID(ctx context.Context, id uint64) (*domain.Dashboard, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *dashboardUsecase) List(ctx context.Context, f dto.DashboardFilter) ([]domain.Dashboard, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *dashboardUsecase) Delete(ctx context.Context, id uint64) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}
	if existing.SystemOwner {
		return domain.ErrSystemOwned
	}
	return u.repo.Delete(ctx, id)
}
