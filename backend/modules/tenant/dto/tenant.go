package dto

import "github.com/utmstack/utmstack/backend/modules/tenant/domain"

type CreateRequest struct {
	Name   string `json:"name"   binding:"required"`
	Domain string `json:"domain" binding:"required"`
}

type UpdateRequest struct {
	Name   string              `json:"name,omitempty"`
	Domain string              `json:"domain,omitempty"`
	Status domain.TenantStatus `json:"status,omitempty"`
}

type Filter struct {
	Name   string              `form:"name"`
	Domain string              `form:"domain"`
	Status domain.TenantStatus `form:"status"`
	Page   int                 `form:"page"`
	Size   int                 `form:"size"`
}
