package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type datasourceUsecase struct {
	repo         connectors.DatasourceRepository
	projector    connectors.AssetProjector
	agentManager *agentmanager.AgentManagerClient
}

func NewDatasourceUsecase(
	repo connectors.DatasourceRepository,
	projector connectors.AssetProjector,
	agentManager *agentmanager.AgentManagerClient,
) connectors.DatasourceUsecase {
	return &datasourceUsecase{
		repo:         repo,
		projector:    projector,
		agentManager: agentManager,
	}
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

func (u *datasourceUsecase) Ping(ctx context.Context, req dto.PingRequest) error {
	now := time.Now().UTC()
	items := make([]domain.Datasource, 0, len(req.Datasources))
	for _, e := range req.Datasources {
		ts := now
		if e.LastPingAt != nil {
			ts = e.LastPingAt.UTC()
		}
		items = append(items, domain.Datasource{
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
	return u.repo.UpsertBatch(ctx, items)
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

func (u *datasourceUsecase) Enrichment(ctx context.Context) ([]dto.DatasourceEnrichment, error) {
	rows, err := u.repo.EnrichmentRows(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]dto.DatasourceEnrichment, 0, len(rows))
	for i := range rows {
		d := &rows[i]
		e := dto.DatasourceEnrichment{
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
	source, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if source == nil {
		return domain.ErrNotFound
	}

	if source.SourceKind == "agent" {
		agentID, err := agentIDFromMetadata(source.Metadata)
		if err != nil {
			return err
		}
		if err := u.agentManager.ForgetAgent(ctx, agentID); err != nil {
			return err
		}
	}

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}
	// The deleted datasource may have carried CIA — rebuild the asset set.
	return u.ProjectAssets(ctx)
}

// ProjectAssets rebuilds tenants.yaml from every datasource with non-zero CIA.
func (u *datasourceUsecase) ProjectAssets(ctx context.Context) error {
	if u.projector == nil {
		return nil
	}
	rows, err := u.repo.ListSensitive(ctx)
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

func agentIDFromMetadata(raw []byte) (uint32, error) {
	if len(raw) == 0 {
		return 0, fmt.Errorf("agent metadata is empty")
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return 0, fmt.Errorf("decode agent metadata: %w", err)
	}
	v, ok := m["agentId"]
	if !ok || v == "" {
		return 0, fmt.Errorf("agent metadata missing agentId")
	}
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse agentId %q: %w", v, err)
	}
	return uint32(n), nil
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
