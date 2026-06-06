package usecase

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	lsErrors "github.com/utmstack/utmstack/backend/modules/logstash/errors"
)

type filterGroupUsecase struct {
	repo connectors.FilterGroupRepository
}

func NewFilterGroupUsecase(repo connectors.FilterGroupRepository) connectors.FilterGroupUsecase {
	return &filterGroupUsecase{repo: repo}
}

func (u *filterGroupUsecase) Create(ctx context.Context, req dto.CreateFilterGroupRequest) (*dto.FilterGroupResponse, error) {
	if req.ID != nil {
		return nil, lsErrors.ErrFilterGroupIDExists
	}

	group := &domain.UtmLogstashFilterGroup{
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
		SystemOwner:      false,
	}
	if err := u.repo.Create(ctx, group); err != nil {
		return nil, err
	}
	return toFilterGroupResponse(group), nil
}

func (u *filterGroupUsecase) Update(ctx context.Context, req dto.UpdateFilterGroupRequest) (*dto.FilterGroupResponse, error) {
	if req.ID == 0 {
		return nil, lsErrors.ErrFilterGroupIDNull
	}

	existing, err := u.repo.GetByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, lsErrors.ErrFilterGroupNotFound
	}

	existing.GroupName = req.GroupName
	existing.GroupDescription = req.GroupDescription

	if err := u.repo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return toFilterGroupResponse(existing), nil
}

func (u *filterGroupUsecase) GetByID(ctx context.Context, id int64) (*dto.FilterGroupResponse, error) {
	group, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, lsErrors.ErrFilterGroupNotFound
	}
	return toFilterGroupResponse(group), nil
}

func (u *filterGroupUsecase) List(ctx context.Context, filters dto.FilterGroupListFilters) ([]dto.FilterGroupResponse, int64, error) {
	groups, total, err := u.repo.List(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.FilterGroupResponse, 0, len(groups))
	for i := range groups {
		result = append(result, *toFilterGroupResponse(&groups[i]))
	}
	return result, total, nil
}

func (u *filterGroupUsecase) Count(ctx context.Context, filters dto.FilterGroupCountFilters) (int64, error) {
	return u.repo.Count(ctx, filters)
}

func (u *filterGroupUsecase) Delete(ctx context.Context, id int64) error {
	return u.repo.Delete(ctx, id)
}

func toFilterGroupResponse(g *domain.UtmLogstashFilterGroup) *dto.FilterGroupResponse {
	return &dto.FilterGroupResponse{
		ID:               g.ID,
		GroupName:        g.GroupName,
		GroupDescription: g.GroupDescription,
		SystemOwner:      g.SystemOwner,
	}
}
