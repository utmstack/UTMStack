package dto

import "github.com/utmstack/utmstack/backend/modules/tenant/domain"

type CreateRequest struct {
	Name       string `json:"name"   binding:"required"`
	Domain     string `json:"domain" binding:"required"`
	AdminEmail string `json:"adminEmail" binding:"required,email"`
	AdminLogin string `json:"adminLogin,omitempty"`
}

type UpdateRequest struct {
	Name   string              `json:"name,omitempty"`
	Domain string              `json:"domain,omitempty"`
	Status domain.TenantStatus `json:"status,omitempty"`

	// A limit left out keeps its value; zero lifts it.
	MaxAIRequests *int `json:"maxAIRequests,omitempty"`
}

type Filter struct {
	Name   string              `form:"name"`
	Domain string              `form:"domain"`
	Status domain.TenantStatus `form:"status"`
	Page   int                 `form:"page"`
	Size   int                 `form:"size"`
}

type SupportAccessRequest struct {
	SupportAccess domain.SupportAccess `json:"supportAccess" binding:"required"`
}
