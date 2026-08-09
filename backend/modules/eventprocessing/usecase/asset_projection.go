package usecase

import (
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type assetProjection struct{ writer *pipelineWriter }

func NewAssetProjection(writer *pipelineWriter) connectors.AssetProjectionUsecase {
	return &assetProjection{writer: writer}
}

func (u *assetProjection) ProjectAssets(assets []common_models.AssetSensitivity) error {
	byTenant := map[string][]assetFileYAML{}
	order := make([]string, 0, 4)
	for _, a := range assets {
		tid := a.TenantID
		if tid == "" {
			tid = DefaultTenantID
		}
		if _, seen := byTenant[tid]; !seen {
			order = append(order, tid)
		}
		byTenant[tid] = append(byTenant[tid], assetFileYAML{
			Name:            a.Name,
			Hostnames:       a.Hostnames,
			Ips:             a.Ips,
			Confidentiality: uint32(a.Confidentiality),
			Integrity:       uint32(a.Integrity),
			Availability:    uint32(a.Availability),
		})
	}

	out := make([]TenantAssets, 0, len(order))
	for _, tid := range order {
		out = append(out, TenantAssets{ID: tid, Assets: byTenant[tid]})
	}
	return u.writer.WriteTenants(out)
}
