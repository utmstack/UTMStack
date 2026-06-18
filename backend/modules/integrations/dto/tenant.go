package dto

import "github.com/utmstack/utmstack/backend/modules/integrations/domain"

// TenantRequest is the create/update body for an integration tenant. Config holds
// the field values; a sensitive field may arrive as MaskedValue to mean "keep
// the stored secret".
type TenantRequest struct {
	Name   string            `json:"name" binding:"required"`
	Config map[string]string `json:"config"`
}

// TenantResponse is the API view of a tenant. Sensitive values are masked (panel)
// or decrypted (plugins) by the usecase before mapping.
type TenantResponse struct {
	Name   string            `json:"name"`
	Config map[string]string `json:"config"`
}

// FromTenant maps a domain tenant (already masked/decrypted) to its response.
func FromTenant(t domain.Tenant) TenantResponse {
	return TenantResponse{Name: t.Name, Config: t.Config}
}
