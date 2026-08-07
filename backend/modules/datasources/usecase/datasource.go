package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type datasourceUsecase struct {
	repo      connectors.DatasourceRepository
	groups    connectors.AssetGroupRepository // nil means UpdateGroup skips cross-tenant validation
	projector connectors.AssetProjector       // may be nil (projection disabled)
}

func NewDatasourceUsecase(repo connectors.DatasourceRepository, groups connectors.AssetGroupRepository, projector connectors.AssetProjector) connectors.DatasourceUsecase {
	return &datasourceUsecase{repo: repo, groups: groups, projector: projector}
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

func (u *datasourceUsecase) Count(ctx context.Context) (int64, error) {
	return u.repo.Count(ctx)
}

// Ping is a batch that spans tenants: one agent-manager serves every tenant's
// agents. The tenant therefore travels per entry, and the write runs unscoped so
// the callback does not collapse the batch onto one of them.
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
			SourceRef:    e.SourceRef,
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

// tenantOf resolves the tenant a batch entry belongs to: what the caller said,
// then the caller's own, then the default. A caller that has not been taught to
// send one keeps working on the single-tenant install it was written for.
func tenantOf(ctx context.Context, entry string) string {
	if entry != "" {
		return entry
	}
	if t := authz.TenantIDFromContext(ctx); t != "" {
		return t
	}
	return authz.DefaultTenantID
}

func (u *datasourceUsecase) Register(ctx context.Context, req dto.RegisterRequest) error {
	now := time.Now().UTC()
	item := domain.Datasource{
		SourceRef:    req.SourceRef,
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

// Enrichment feeds the alert plugin cache, which needs every tenant's rows
// (each carries its own TenantID). Scoped to the caller's tenant it would
// return only their slice and quietly drop the rest.
func (u *datasourceUsecase) Enrichment(ctx context.Context) ([]dto.DatasourceEnrichment, error) {
	rows, err := u.repo.EnrichmentRows(tenancy.WithAllTenantsRead(ctx))
	if err != nil {
		return nil, err
	}
	out := make([]dto.DatasourceEnrichment, 0, len(rows))
	for i := range rows {
		d := &rows[i]
		e := dto.DatasourceEnrichment{
			TenantID: d.TenantID,
			Name:     d.Name,
			DataType: d.DataType,
			GroupID:  d.GroupID,
			Labels:   splitLabels(d.Labels),
		}
		if d.Group != nil {
			e.GroupName = d.Group.GroupName
		}
		out = append(out, e)
	}
	return out, nil
}

// splitLabels turns the stored comma-separated labels into a trimmed slice.
func splitLabels(labels string) []string {
	if strings.TrimSpace(labels) == "" {
		return nil
	}
	parts := strings.Split(labels, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func (u *datasourceUsecase) UpdateGroup(ctx context.Context, req dto.UpdateGroupRequest) error {
	// Cross-tenant guard: the datasource update itself is scoped by the tenancy
	// callback, but the group_id is a raw uint64 from the client — nothing stops
	// tenant A from pointing their datasources at tenant B's group. The group
	// lookup is tenant-scoped by the same callback, so a miss means it either
	// does not exist or belongs to someone else.
	if req.GroupID != nil && u.groups != nil {
		g, err := u.groups.FindByID(ctx, *req.GroupID)
		if err != nil {
			return err
		}
		if g == nil {
			return domain.ErrNotFound
		}
	}
	return u.repo.UpdateGroup(ctx, req.IDs, req.GroupID)
}

func (u *datasourceUsecase) UpdateLabels(ctx context.Context, req dto.UpdateLabelsRequest) error {
	return u.repo.UpdateLabels(ctx, req.ID, req.Labels)
}

func (u *datasourceUsecase) UpdateSensitivity(ctx context.Context, req dto.UpdateSensitivityRequest) error {
	if err := u.repo.UpdateSensitivity(ctx, req.ID,
		clampCIA(req.AssetConfidentiality), clampCIA(req.AssetIntegrity), clampCIA(req.AssetAvailability)); err != nil {
		return err
	}
	return u.ProjectAssets(ctx)
}

func (u *datasourceUsecase) Delete(ctx context.Context, id uint64) error {
	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	// The deleted datasource may have carried CIA — rebuild the asset set.
	return u.ProjectAssets(ctx)
}

// ProjectAssets rebuilds tenants.yaml from every datasource with non-zero CIA.
// The read spans tenants because the projector writes one shared file: scoped
// to the caller's tenant, an UpdateSensitivity or Delete on tenant A would
// rewrite the file with only A's assets and wipe every other tenant's.
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
