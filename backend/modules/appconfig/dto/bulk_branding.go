package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type BulkBrandingUpdateRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Branding BrandingRequest                  `json:"branding"`
}

type BulkBrandingAssetRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
}
