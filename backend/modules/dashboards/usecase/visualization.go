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

func (u *visualizationUsecase) Create(ctx context.Context, v *domain.UtmVisualization, user string) (*domain.UtmVisualization, error) {
	if v.ID != 0 {
		return nil, domain.ErrIDForbidden
	}
	if strings.TrimSpace(v.Name) == "" {
		return nil, domain.ErrNameRequired
	}
	now := time.Now().UTC()
	v.CreatedDate = now
	v.ModifiedDate = now
	v.UserCreated = user
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) Update(ctx context.Context, v *domain.UtmVisualization, user string) (*domain.UtmVisualization, error) {
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
	v.CreatedDate = existing.CreatedDate
	v.UserCreated = existing.UserCreated
	v.SystemOwner = existing.SystemOwner
	v.ModifiedDate = time.Now().UTC()
	v.UserModified = user
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) GetByID(ctx context.Context, id uint64) (*domain.UtmVisualization, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *visualizationUsecase) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.UtmVisualization, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *visualizationUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
