package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type DashboardUsecase interface {
	Create(ctx context.Context, d *domain.Dashboard, user string) (*domain.Dashboard, error)
	Update(ctx context.Context, d *domain.Dashboard, user string) (*domain.Dashboard, error)
	GetByID(ctx context.Context, id uint64) (*domain.Dashboard, error)
	List(ctx context.Context, f dto.DashboardFilter) ([]domain.Dashboard, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type VisualizationUsecase interface {
	Create(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error)
	Update(ctx context.Context, v *domain.Visualization, user string) (*domain.Visualization, error)
	GetByID(ctx context.Context, id uint64) (*domain.Visualization, error)
	List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}
