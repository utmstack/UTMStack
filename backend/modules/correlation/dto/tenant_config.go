package dto

import (
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
)

type CreateTenantConfigRequest struct {
	ID                   *int64          `json:"id"`
	AssetName            string          `json:"assetName"`
	AssetHostnameListDef json.RawMessage `json:"assetHostnameListDef"`
	AssetIpListDef       json.RawMessage `json:"assetIpListDef"`
	AssetConfidentiality int             `json:"assetConfidentiality"`
	AssetIntegrity       int             `json:"assetIntegrity"`
	AssetAvailability    int             `json:"assetAvailability"`
}

type UpdateTenantConfigRequest struct {
	ID                   *int64          `json:"id"`
	AssetName            string          `json:"assetName"`
	AssetHostnameListDef json.RawMessage `json:"assetHostnameListDef"`
	AssetIpListDef       json.RawMessage `json:"assetIpListDef"`
	AssetConfidentiality int             `json:"assetConfidentiality"`
	AssetIntegrity       int             `json:"assetIntegrity"`
	AssetAvailability    int             `json:"assetAvailability"`
}

type TenantConfigResponse struct {
	ID                   int64           `json:"id"`
	AssetName            string          `json:"assetName"`
	AssetHostnameListDef json.RawMessage `json:"assetHostnameListDef"`
	AssetIpListDef       json.RawMessage `json:"assetIpListDef"`
	AssetConfidentiality int             `json:"assetConfidentiality"`
	AssetIntegrity       int             `json:"assetIntegrity"`
	AssetAvailability    int             `json:"assetAvailability"`
	LastUpdate           *time.Time      `json:"lastUpdate"`
}

type TenantConfigFilters struct {
	// Page is 0-based.
	Page   int    `form:"page"`
	Size   int    `form:"size"`
	Search string `form:"search"`
}

func rawOrNull(s string) json.RawMessage {
	if s == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(s)
}

func stringOrEmpty(r json.RawMessage) string {
	if len(r) == 0 {
		return ""
	}
	return string(r)
}

func TenantConfigToResponse(e *domain.UtmTenantConfig) *TenantConfigResponse {
	return &TenantConfigResponse{
		ID:                   e.ID,
		AssetName:            e.AssetName,
		AssetHostnameListDef: rawOrNull(e.AssetHostnameListDef),
		AssetIpListDef:       rawOrNull(e.AssetIpListDef),
		AssetConfidentiality: e.AssetConfidentiality,
		AssetIntegrity:       e.AssetIntegrity,
		AssetAvailability:    e.AssetAvailability,
		LastUpdate:           e.LastUpdate,
	}
}
