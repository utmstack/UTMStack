package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type layoutUsecase struct{ repo connectors.LayoutRepository }

func NewLayoutUsecase(repo connectors.LayoutRepository) connectors.LayoutUsecase {
	return &layoutUsecase{repo: repo}
}

func (u *layoutUsecase) Create(ctx context.Context, l *domain.UtmDashboardVisualization) (*domain.UtmDashboardVisualization, error) {
	if l.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if err := u.repo.Save(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (u *layoutUsecase) Update(ctx context.Context, l *domain.UtmDashboardVisualization) (*domain.UtmDashboardVisualization, error) {
	if l.ID == 0 {
		return nil, domain.ErrIDRequired
	}
	existing, err := u.repo.FindByID(ctx, l.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	if err := u.repo.Save(ctx, l); err != nil {
		return nil, err
	}
	return l, nil
}

func (u *layoutUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmDashboardVisualization, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *layoutUsecase) List(ctx context.Context, f dto.LayoutFilter) ([]domain.UtmDashboardVisualization, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *layoutUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
