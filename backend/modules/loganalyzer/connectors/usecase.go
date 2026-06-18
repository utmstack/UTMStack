package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type QueryUsecase interface {
	Create(ctx context.Context, q *domain.UtmLogAnalyzerQuery, owner string) (*domain.UtmLogAnalyzerQuery, error)
	Update(ctx context.Context, q *domain.UtmLogAnalyzerQuery, owner string) (*domain.UtmLogAnalyzerQuery, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmLogAnalyzerQuery, error)
	List(ctx context.Context, f dto.QueryFilter) ([]domain.UtmLogAnalyzerQuery, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type AnalyzerUsecase interface {
	TopValues(ctx context.Context, index, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error)
	ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error)
}
