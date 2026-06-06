package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	lsErrors "github.com/utmstack/utmstack/backend/modules/logstash/errors"
)

type filterUsecase struct {
	repo               connectors.FilterRepository
	pipelineFilterRepo connectors.PipelineFilterRepository
	pipelineRepo       connectors.PipelineRepository
}

func NewFilterUsecase(
	repo connectors.FilterRepository,
	pipelineFilterRepo connectors.PipelineFilterRepository,
	pipelineRepo connectors.PipelineRepository,
) connectors.FilterUsecase {
	return &filterUsecase{
		repo:               repo,
		pipelineFilterRepo: pipelineFilterRepo,
		pipelineRepo:       pipelineRepo,
	}
}

func (u *filterUsecase) Create(ctx context.Context, req dto.CreateFilterRequest, pipelineID *int64) (*dto.FilterResponse, error) {
	if req.ID != nil && *req.ID != 0 {
		return nil, lsErrors.ErrFilterIDExists
	}

	now := time.Now().UTC()
	filter := &domain.UtmLogstashFilter{
		FilterName:     req.FilterName,
		LogstashFilter: req.LogstashFilter,
		FilterGroupID:  req.FilterGroupID,
		DataTypeID:     req.DataTypeID,
		SystemOwner:    req.SystemOwner,
		IsActive:       req.IsActive,
		ModuleName:     req.ModuleName,
		FilterVersion:  req.FilterVersion,
		UpdatedAt:      &now,
	}

	if err := u.repo.Create(ctx, filter); err != nil {
		return nil, err
	}

	if pipelineID != nil {
		pipeline, err := u.pipelineRepo.GetByID(ctx, *pipelineID)
		if err != nil {
			return nil, err
		}
		if pipeline == nil {
			return nil, lsErrors.ErrPipelineNotFound
		}

		rel := &domain.UtmGroupLogstashPipelineFilters{
			FilterID:   int32(filter.ID),
			PipelineID: int32(*pipelineID),
			Relation:   domain.RelationUserCustomFilter,
		}
		if err := u.pipelineFilterRepo.Save(ctx, rel); err != nil {
			return nil, err
		}
	}

	return toFilterResponse(filter), nil
}

func (u *filterUsecase) Update(ctx context.Context, req dto.UpdateFilterRequest) (*dto.FilterResponse, error) {
	if req.ID == 0 {
		return nil, lsErrors.ErrFilterIDNull
	}
	if req.SystemOwner {
		return nil, lsErrors.ErrFilterSystemOwner
	}

	now := time.Now().UTC()
	filter := &domain.UtmLogstashFilter{
		ID:             req.ID,
		FilterName:     req.FilterName,
		LogstashFilter: req.LogstashFilter,
		FilterGroupID:  req.FilterGroupID,
		DataTypeID:     req.DataTypeID,
		SystemOwner:    req.SystemOwner,
		IsActive:       req.IsActive,
		ModuleName:     req.ModuleName,
		FilterVersion:  req.FilterVersion,
		UpdatedAt:      &now,
	}

	if err := u.repo.Update(ctx, filter); err != nil {
		return nil, err
	}
	return toFilterResponse(filter), nil
}

func (u *filterUsecase) GetByID(ctx context.Context, id int64) (*dto.FilterResponse, error) {
	filter, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if filter == nil {
		return nil, lsErrors.ErrFilterNotFound
	}
	return toFilterResponse(filter), nil
}

func (u *filterUsecase) List(ctx context.Context, filters dto.FilterFilters) ([]dto.FilterResponse, int64, error) {
	items, total, err := u.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}
	result := make([]dto.FilterResponse, 0, len(items))
	for i := range items {
		result = append(result, *toFilterResponse(&items[i]))
	}
	return result, total, nil
}

func (u *filterUsecase) FiltersByPipelineID(ctx context.Context, pipelineID int64) ([]dto.FilterResponse, error) {
	items, err := u.repo.FiltersByPipelineID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.FilterResponse, 0, len(items))
	for i := range items {
		result = append(result, *toFilterResponse(&items[i]))
	}
	return result, nil
}

func (u *filterUsecase) Delete(ctx context.Context, id int64) error {
	if err := u.pipelineFilterRepo.DeleteByFilterID(ctx, int32(id)); err != nil {
		return err
	}
	return u.repo.Delete(ctx, id)
}

func toFilterResponse(f *domain.UtmLogstashFilter) *dto.FilterResponse {
	return &dto.FilterResponse{
		ID:             f.ID,
		FilterName:     f.FilterName,
		LogstashFilter: f.LogstashFilter,
		FilterGroupID:  f.FilterGroupID,
		SystemOwner:    f.SystemOwner,
		IsActive:       f.IsActive,
		ModuleName:     f.ModuleName,
		FilterVersion:  f.FilterVersion,
		UpdatedAt:      f.UpdatedAt,
	}
}
