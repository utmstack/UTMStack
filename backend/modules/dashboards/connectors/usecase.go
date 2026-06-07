package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
)

type DashboardUsecase interface {
	Create(ctx context.Context, d *domain.UtmDashboard, user string) (*domain.UtmDashboard, error)
	Update(ctx context.Context, d *domain.UtmDashboard, user string) (*domain.UtmDashboard, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmDashboard, error)
	List(ctx context.Context, f dto.DashboardFilter) ([]domain.UtmDashboard, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type VisualizationUsecase interface {
	Create(ctx context.Context, v *domain.UtmVisualization, user string) (*domain.UtmVisualization, error)
	Update(ctx context.Context, v *domain.UtmVisualization, user string) (*domain.UtmVisualization, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmVisualization, error)
	List(ctx context.Context, f dto.VisualizationFilter) ([]domain.UtmVisualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type LayoutUsecase interface {
	Create(ctx context.Context, l *domain.UtmDashboardVisualization) (*domain.UtmDashboardVisualization, error)
	Update(ctx context.Context, l *domain.UtmDashboardVisualization) (*domain.UtmDashboardVisualization, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmDashboardVisualization, error)
	List(ctx context.Context, f dto.LayoutFilter) ([]domain.UtmDashboardVisualization, int64, error)
	Delete(ctx context.Context, id uint64) error
}
