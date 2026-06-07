package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type datasourceUsecase struct {
	repo connectors.DatasourceRepository
}

func NewDatasourceUsecase(repo connectors.DatasourceRepository) connectors.DatasourceUsecase {
	return &datasourceUsecase{repo: repo}
}

func (u *datasourceUsecase) GetByID(ctx context.Context, id uint64) (*dto.DatasourceDTO, error) {
	d, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, domain.ErrNotFound
	}
	out := dto.ToDatasourceDTO(d)
	return &out, nil
}

func (u *datasourceUsecase) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[dto.DatasourceDTO], error) {
	res, err := u.repo.List(ctx, req)
	if err != nil {
		return common_models.ListResponse[dto.DatasourceDTO]{}, err
	}
	return common_models.MapList(res, func(d domain.Datasource) dto.DatasourceDTO {
		return dto.ToDatasourceDTO(&d)
	}), nil
}

func (u *datasourceUsecase) Ping(ctx context.Context, req dto.PingRequest) error {
	now := time.Now().UTC()
	items := make([]domain.Datasource, 0, len(req.Datasources))
	for _, e := range req.Datasources {
		ts := now
		if e.LastPingAt != nil {
			ts = e.LastPingAt.UTC()
		}
		items = append(items, domain.Datasource{
			Name:         e.Name,
			SourceKind:   e.SourceKind,
			IP:           e.IP,
			Metadata:     e.Metadata,
			LastPingAt:   &ts,
			ModifiedAt:   &ts,
			DiscoveredAt: &ts, // only applied on insert; preserved on update
		})
	}
	return u.repo.UpsertBatch(ctx, items)
}

func (u *datasourceUsecase) UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error {
	return u.repo.UpdateGroup(ctx, req.IDs, req.GroupID)
}

func (u *datasourceUsecase) UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error {
	return u.repo.UpdateLabels(ctx, req.ID, req.Labels)
}

func (u *datasourceUsecase) Delete(ctx context.Context, id uint64) error {
	return u.repo.Delete(ctx, id)
}
