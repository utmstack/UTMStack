package dto

import (
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type BulkCreateFrameworkRequest struct {
	Selector  common_models.BulkTenantSelector `json:"selector"  binding:"required"`
	Framework domain.Framework                 `json:"framework" binding:"required"`
}

type BulkUpdateFrameworkRequest struct {
	Selector  common_models.BulkTenantSelector `json:"selector"  binding:"required"`
	Framework domain.Framework                 `json:"framework" binding:"required"`
}

type BulkDeleteFrameworkRequest struct {
	Selector    common_models.BulkTenantSelector `json:"selector"     binding:"required"`
	FrameworkKey string                           `json:"frameworkKey" binding:"required"`
}

type BulkCreateControlRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector" binding:"required"`
	Control  domain.Control                   `json:"control"  binding:"required"`
}

type BulkUpdateControlRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector" binding:"required"`
	Control  domain.Control                   `json:"control"  binding:"required"`
}

type BulkDeleteControlRequest struct {
	Selector  common_models.BulkTenantSelector `json:"selector"   binding:"required"`
	ControlID string                           `json:"controlId"  binding:"required"`
}
