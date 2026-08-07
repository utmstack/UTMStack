package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/repository"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type analyzerUsecase struct{ repo connectors.AnalyzerRepository }

func NewAnalyzerUsecase(repo connectors.AnalyzerRepository) connectors.AnalyzerUsecase {
	return &analyzerUsecase{repo: repo}
}

func (u *analyzerUsecase) TopValues(ctx context.Context, dataset, dataType, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error) {
	if dataset == "" {
		return nil, domain.ErrDatasetRequired
	}
	if field == "" {
		return nil, domain.ErrFieldRequired
	}
	return u.repo.TopValues(ctx, dataset, dataType, field, filters, top)
}

func (u *analyzerUsecase) ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error) {
	if req.Dataset == "" {
		return nil, domain.ErrDatasetRequired
	}
	if req.Field == "" {
		return nil, domain.ErrFieldRequired
	}
	return u.repo.ChartView(ctx, req)
}

func (u *analyzerUsecase) Datasets() []string { return repository.Datasets() }

func (u *analyzerUsecase) Fields(ctx context.Context, dataset string) ([]dto.Field, error) {
	return u.repo.Fields(ctx, dataset)
}

func (u *analyzerUsecase) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	return u.repo.Search(ctx, req)
}

func (u *analyzerUsecase) DataTypes(ctx context.Context, dataset string) ([]string, error) {
	return u.repo.DataTypes(ctx, dataset)
}

func (u *analyzerUsecase) SearchSQL(ctx context.Context, sql string, page, size int) (*dto.SearchResponse, error) {
	if err := ValidateSQL(sql); err != nil {
		return nil, fmt.Errorf("%w: %s", domain.ErrInvalidSQL, err.Error())
	}
	return u.repo.SearchSQL(ctx, sql, page, size)
}
