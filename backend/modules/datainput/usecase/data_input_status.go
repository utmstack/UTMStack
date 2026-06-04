package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/datainput/connectors"
	"github.com/utmstack/utmstack/backend/modules/datainput/domain"
	"github.com/utmstack/utmstack/backend/modules/datainput/dto"
	diErrors "github.com/utmstack/utmstack/backend/modules/datainput/errors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
)

const defaultMedian int64 = 86400

type dataInputStatusUsecase struct {
	repo connectors.DataInputStatusRepository
}

func NewDataInputStatusUsecase(repo connectors.DataInputStatusRepository) connectors.DataInputStatusUsecase {
	return &dataInputStatusUsecase{repo: repo}
}

func (u *dataInputStatusUsecase) Create(ctx context.Context, req dto.CreateDataInputStatusRequest) (*dto.DataInputStatusResponse, error) {
	if req.ID != nil && *req.ID != "" {
		return nil, diErrors.ErrIDExists
	}

	median := req.Median
	if median == nil {
		m := defaultMedian
		median = &m
	}

	entity := &domain.UtmDataInputStatus{
		// ID is set by the repository before persisting.
		Source:    req.Source,
		DataType:  req.DataType,
		Timestamp: req.Timestamp,
		Median:    median,
		Alias:     req.Alias,
	}

	if err := u.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	return toResponse(entity), nil
}

func (u *dataInputStatusUsecase) Update(ctx context.Context, req dto.UpdateDataInputStatusRequest) (*dto.DataInputStatusResponse, error) {
	if req.ID == "" {
		return nil, diErrors.ErrIDNull
	}

	entity := &domain.UtmDataInputStatus{
		ID:        req.ID,
		Source:    req.Source,
		DataType:  req.DataType,
		Timestamp: req.Timestamp,
		Median:    req.Median,
		Alias:     req.Alias,
	}

	if err := u.repo.Update(ctx, entity); err != nil {
		return nil, err
	}
	return toResponse(entity), nil
}

func (u *dataInputStatusUsecase) GetByID(ctx context.Context, id string) (*dto.DataInputStatusResponse, error) {
	entity, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, diErrors.ErrDataInputStatusNotFound
	}
	return toResponse(entity), nil
}

func (u *dataInputStatusUsecase) List(ctx context.Context, filters dto.DataInputStatusListFilters) ([]dto.DataInputStatusResponse, int64, error) {
	page, size := normPage(filters.Page, filters.Size)

	items, total, err := u.repo.FindImportantDatasource(ctx, page, size)
	if err != nil {
		logger.Warn("datainput: FindImportantDatasource failed — returning null body: " + err.Error())
		return nil, 0, nil
	}

	responses := make([]dto.DataInputStatusResponse, 0, len(items))
	for i := range items {
		responses = append(responses, *toResponse(&items[i]))
	}
	return responses, total, nil
}

func (u *dataInputStatusUsecase) Count(ctx context.Context, filters dto.DataInputStatusCountFilters) (int64, error) {
	return u.repo.Count(ctx, filters)
}

func (u *dataInputStatusUsecase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func normPage(page, size int) (int, int) {
	if size <= 0 {
		size = 20
	}
	if page < 0 {
		page = 0
	}
	return page, size
}

func toResponse(e *domain.UtmDataInputStatus) *dto.DataInputStatusResponse {
	return &dto.DataInputStatusResponse{
		ID:        e.ID,
		Source:    e.Source,
		DataType:  e.DataType,
		Timestamp: e.Timestamp,
		Median:    e.Median,
		Alias:     e.Alias,
		IsDown:    e.IsDown(),
	}
}
