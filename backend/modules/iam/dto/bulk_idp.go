package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type BulkCreateIDPRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Provider IdentityProviderRequest           `json:"provider"`
}

type BulkUpdateIDPRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Provider IdentityProviderRequest           `json:"provider"`
}

type BulkDeleteIDPRequest struct {
	Selector   common_models.BulkTenantSelector `json:"selector"`
	ProviderID string                           `json:"providerId"`
}
