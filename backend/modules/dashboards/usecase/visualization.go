package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/dashboards/connectors"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"

	"github.com/google/uuid"
)

type visualizationUsecase struct {
	repo connectors.VisualizationRepository
}

func NewVisualizationUsecase(repo connectors.VisualizationRepository) connectors.VisualizationUsecase {
	return &visualizationUsecase{repo: repo}
}

func (u *visualizationUsecase) Create(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error) {
	if v.ID != uuid.Nil {
		return nil, domain.ErrIDForbidden
	}
	if v.DashboardID == uuid.Nil {
		return nil, domain.ErrDashboardIDRequired
	}
	if err := sanitizeVisualization(v); err != nil {
		return nil, err
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
	if existing.SystemOwner {
		return nil, domain.ErrSystemOwned
	}
	if err := sanitizeVisualization(v); err != nil {
		return nil, err
	}
	v.CreatedDate = existing.CreatedDate
	v.SystemOwner = existing.SystemOwner
	// A visualization can't move to a different dashboard — it's not reusable.
	v.DashboardID = existing.DashboardID
	v.ModifiedDate = time.Now().UTC()
	if err := u.repo.Save(ctx, v); err != nil {
		return nil, err
	}
	return v, nil
}

func (u *visualizationUsecase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Visualization, error) {
	return u.repo.FindByID(ctx, id)
}

func (u *visualizationUsecase) List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error) {
	return u.repo.List(ctx, f)
}

func (u *visualizationUsecase) Delete(ctx context.Context, id uuid.UUID) error {
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

func sanitizeVisualization(v *domain.Visualization) error {
	if strings.TrimSpace(v.Spec) == "" {
		return domain.ErrSpecRequired
	}

	var spec domain.Spec
	if err := json.Unmarshal([]byte(v.Spec), &spec); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidSpec, err.Error())
	}
	if err := spec.Validate(); err != nil {
		return fmt.Errorf("%w: %s", domain.ErrInvalidSpec, err.Error())
	}
	return nil
}
