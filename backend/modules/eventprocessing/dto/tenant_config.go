package dto

// Identity for tenant config (assets) is assetName (string). No int64 ID.

type CreateTenantConfigRequest struct {
	AssetName            string   `json:"assetName"            binding:"required"`
	AssetHostnameList    []string `json:"assetHostnameList"`
	AssetIpList          []string `json:"assetIpList"`
	AssetConfidentiality int      `json:"assetConfidentiality"`
	AssetIntegrity       int      `json:"assetIntegrity"`
	AssetAvailability    int      `json:"assetAvailability"`
}

type UpdateTenantConfigRequest struct {
	AssetName            string   `json:"assetName"            binding:"required"`
	AssetHostnameList    []string `json:"assetHostnameList"`
	AssetIpList          []string `json:"assetIpList"`
	AssetConfidentiality int      `json:"assetConfidentiality"`
	AssetIntegrity       int      `json:"assetIntegrity"`
	AssetAvailability    int      `json:"assetAvailability"`
}

type TenantConfigResponse struct {
	AssetName            string   `json:"assetName"`
	AssetHostnameList    []string `json:"assetHostnameList"`
	AssetIpList          []string `json:"assetIpList"`
	AssetConfidentiality int      `json:"assetConfidentiality"`
	AssetIntegrity       int      `json:"assetIntegrity"`
	AssetAvailability    int      `json:"assetAvailability"`
}

type TenantConfigFilters struct {
	Search string `form:"search"`
	Page   int    `form:"page"`
	Size   int    `form:"size"`
}
