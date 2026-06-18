package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type visualizationUsecase struct {
	repo connectors.VisualizationRepository
}

func NewVisualizationUsecase(repo connectors.VisualizationRepository) connectors.VisualizationUsecase {
	return &visualizationUsecase{repo: repo}
}

func (u *visualizationUsecase) Create(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error) {
	if v.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(v.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	if strings.TrimSpace(v.SQLQuery) == "" {
		return nil, domain.ErrSQLQueryRequired
	}
	now := time.Now().UTC()
	v.CreatedDate = now
	v.ModifiedDate = now
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) Update(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error) {
	if v.ID == 0 {
		return nil, domain.ErrIDRequired
	}
	existing, err := u.repo.FindByID(ctx, v.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}
	if strings.TrimSpace(v.SQLQuery) == "" {
		return nil, domain.ErrSQLQueryRequired
	}
	v.CreatedDate = existing.CreatedDate
	v.SystemOwner = existing.SystemOwner
	v.ModifiedDate = time.Now().UTC()
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) GetByID(ctx context.Context, id uint64) (*domain.Visualization, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *visualizationUsecase) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *visualizationUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
