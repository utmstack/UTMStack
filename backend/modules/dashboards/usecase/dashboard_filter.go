package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"

	"github.com/google/uuid"
)

type dashboardFilterUsecase struct {
	repo connectors.DashboardFilterRepository
}

func NewDashboardFilterUsecase(repo connectors.DashboardFilterRepository) connectors.DashboardFilterUsecase {
	return &dashboardFilterUsecase{repo: repo}
}

func (u *dashboardFilterUsecase) Create(ctx context.Context, v *domain.DashboardFilter, user string) (*domain.DashboardFilter, error) {
	if v.ID != uuid.Nil {
		return nil, domain.ErrIDForbidden
	}
	if v.DashboardID == uuid.Nil {
		return nil, domain.ErrDashboardIDRequired
	}

	now := time.Now().UTC()
	v.CreatedDate = now
	v.ModifiedDate = now
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *dashboardFilterUsecase) Update(ctx context.Context, v *domain.DashboardFilter, user string) (*domain.DashboardFilter, error) {
	if v.ID == uuid.Nil {
		return nil, domain.ErrIDRequired
	}

	existing, err := u.repo.FindByID(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	v.CreatedDate = existing.CreatedDate
	// A dashboardFilter can't move to a different tenant — it's not reusable.
	v.TenantID = existing.TenantID
	// A dashboardFilter can't move to a different dashboard — it's not reusable.
	v.DashboardID = existing.DashboardID
	v.ModifiedDate = time.Now().UTC()
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *dashboardFilterUsecase) List(ctx context.Context, f dto.DashboardFilterFilter) ([]domain.DashboardFilter, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *dashboardFilterUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	existing, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}
	return u.repo.Delete(ctx, id)
}

