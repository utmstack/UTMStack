package connectors

import (
	"context"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type QueryRepository interface {
	Save(ctx context.Context, q *domain.SavedQuery) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.SavedQuery, error)
	List(ctx context.Context, f dto.QueryFilter) ([]domain.SavedQuery, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AnalyzerRepository interface {
	SearchSQL(ctx context.Context, sql string, page, size int) (*dto.SearchResponse, error)
	TopValues(ctx context.Context, dataset, dataType, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error)
	ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error)
	Fields(ctx context.Context, dataset string) ([]dto.Field, error)
	Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
	DataTypes(ctx context.Context, dataset string) ([]string, error)
}
