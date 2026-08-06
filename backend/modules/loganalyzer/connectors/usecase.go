package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type QueryUsecase interface {
	Create(ctx context.Context, q *domain.SavedQuery, owner string) (*domain.SavedQuery, error)
	Update(ctx context.Context, q *domain.SavedQuery, owner string) (*domain.SavedQuery, error)
	GetByID(ctx context.Context, id uint64) (*domain.SavedQuery, error)
	List(ctx context.Context, f dto.QueryFilter) ([]domain.SavedQuery, int64, error)
	Delete(ctx context.Context, id uint64) error
}

type AnalyzerUsecase interface {
	TopValues(ctx context.Context, dataset, dataType, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error)
	ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error)
	Fields(ctx context.Context, dataset string) ([]dto.Field, error)
	Datasets() []string
	Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
	DataTypes(ctx context.Context, dataset string) ([]string, error)
}
