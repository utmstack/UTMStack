package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type DashboardRepository interface {
	Save(ctx context.Context, d *domain.Dashboard) error
	FindByID(ctx context.Context, id uint64) (*domain.Dashboard, error)
	List(ctx context.Context, f dto.DashboardFilter) ([]domain.Dashboard, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type VisualizationRepository interface {
	Save(ctx context.Context, v *domain.Visualization) error
	FindByID(ctx context.Context, id uint64) (*domain.Visualization, error)
	List(ctx context.Context, f dto.VisualizationFilter) ([]domain.Visualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}
