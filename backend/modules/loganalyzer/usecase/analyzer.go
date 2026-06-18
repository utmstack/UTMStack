package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type analyzerUsecase struct{ repo connectors.AnalyzerRepository }

func NewAnalyzerUsecase(repo connectors.AnalyzerRepository) connectors.AnalyzerUsecase {
	return &analyzerUsecase{repo: repo}
}

func (u *analyzerUsecase) TopValues(ctx context.Context, index, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error) {
	if index == "" {
		return nil, domain.ErrIndexRequired
	}
	if field == "" {
		return nil, domain.ErrFieldRequired
	}
	return u.repo.TopValues(ctx, index, field, filters, top)
}

func (u *analyzerUsecase) ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error) {
	if req.IndexPattern == "" {
		return nil, domain.ErrIndexRequired
	}
	if req.Field == "" {
		return nil, domain.ErrFieldRequired
	}
	return u.repo.ChartView(ctx, req)
}
