package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var defaultTenant = uuid.MustParse(authz.DefaultTenantID)

type datasourceUsecase struct {
	repo      connectors.DatasourceRepository
	projector connectors.AssetProjector // may be nil (projection disabled)
	notifier  Notifier                  // nil → project inline
}

func NewDatasourceUsecase(repo connectors.DatasourceRepository, projector connectors.AssetProjector) connectors.DatasourceUsecase {
	return &datasourceUsecase{repo: repo, projector: projector}
}

func (u *datasourceUsecase) SetAssetNotifier(n Notifier) { u.notifier = n }

func (u *datasourceUsecase) assetsChanged(ctx context.Context) error {
	if u.notifier != nil {
		u.notifier.Notify()
		return nil
	}
	return u.ProjectAssets(ctx)
}

func (u *datasourceUsecase) GetByID(ctx context.Context, id uuid.UUID) (*dto.DatasourceDTO, error) {
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

func (u *datasourceUsecase) Count(ctx context.Context) (int64, error) {
	return u.repo.Count(ctx)
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
			TenantID:     tenantOf(ctx, e.TenantID),
			Name:         e.Name,
			DataType:     e.DataType,
			SourceKind:   e.SourceKind,
			IP:           e.IP,
			Metadata:     e.Metadata,
			LastPingAt:   &ts,
			ModifiedAt:   &ts,
			DiscoveredAt: &ts, // only applied on insert; preserved on update
		})
	}
	return u.repo.UpsertBatch(tenancy.WithAllTenants(ctx), items)
}

func tenantOf(ctx context.Context, entry uuid.UUID) uuid.UUID {
	if entry != uuid.Nil {
		return entry
	}
	if own, err := uuid.Parse(authz.TenantIDFromContext(ctx)); err == nil {
		return own
	}
	return defaultTenant
}

func (u *datasourceUsecase) Register(ctx context.Context, req dto.RegisterRequest) error {
	now := time.Now().UTC()
	item := domain.Datasource{
		Name:         req.Name,
		DataType:     req.DataType,
		SourceKind:   req.SourceKind,
		IP:           req.IP,
		Metadata:     req.Metadata,
		ModifiedAt:   &now,
		DiscoveredAt: &now, // only applied on insert; preserved on update
	}
	return u.repo.RegisterBatch(ctx, []domain.Datasource{item})
}

func (u *datasourceUsecase) UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error {
	return u.repo.UpdateLabels(ctx, req.ID, req.Labels)
}

func (u *datasourceUsecase) UpdateSensitivity(ctx context.Context, req dto.UpdateSensitivityRequest) error {
	if err := u.repo.UpdateSensitivity(ctx, req.ID,
		clampCIA(req.AssetConfidentiality), clampCIA(req.AssetIntegrity), clampCIA(req.AssetAvailability)); err != nil {
		return err
	}
	return u.assetsChanged(ctx)
}

func (u *datasourceUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	// The deleted datasource may have carried CIA — rebuild the asset set.
	return u.assetsChanged(ctx)
}

func (u *datasourceUsecase) ProjectAssets(ctx context.Context) error {
	if u.projector == nil {
		return nil
	}
	rows, err := u.repo.ListSensitive(tenancy.WithAllTenants(ctx))
	if err != nil {
		return err
	}
	assets := make([]common_models.AssetSensitivity, 0, len(rows))
	for i := range rows {
		d := &rows[i]
		a := common_models.AssetSensitivity{
			TenantID:        d.TenantID.String(),
			Name:            d.Name,
			Hostnames:       []string{d.Name},
			Confidentiality: d.AssetConfidentiality,
			Integrity:       d.AssetIntegrity,
			Availability:    d.AssetAvailability,
		}
		if strings.TrimSpace(d.IP) != "" {
			a.Ips = []string{d.IP}
		}
		assets = append(assets, a)
	}
	return u.projector.ProjectAssets(assets)
}

// clampCIA bounds a sensitivity axis to the valid 0–3 range.
func clampCIA(v int) int {
	if v < 0 {
		return 0
	}
	if v > 3 {
		return 3
	}
	return v
}
