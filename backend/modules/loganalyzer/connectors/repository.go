package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type QueryRepository interface {
	Save(ctx context.Context, q *domain.UtmLogAnalyzerQuery) error
	FindByID(ctx context.Context, id uint64) (*domain.UtmLogAnalyzerQuery, error)
	List(ctx context.Context, f dto.QueryFilter) ([]domain.UtmLogAnalyzerQuery, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type AnalyzerRepository interface {
	TopValues(ctx context.Context, index, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error)
	ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error)
}
