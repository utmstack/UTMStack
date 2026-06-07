package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type DashboardRepository interface {
	Save(ctx context.Context, d *domain.UtmDashboard) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmDashboard, error)
	List(ctx context.Context, f dto.DashboardFilter) ([]domain.UtmDashboard, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type VisualizationRepository interface {
	Save(ctx context.Context, v *domain.UtmVisualization) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmVisualization, error)
	List(ctx context.Context, f dto.VisualizationFilter) ([]domain.UtmVisualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type LayoutRepository interface {
	Save(ctx context.Context, l *domain.UtmDashboardVisualization) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmDashboardVisualization, error)
	List(ctx context.Context, f dto.LayoutFilter) ([]domain.UtmDashboardVisualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}
