package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type BulkCreatePipelineRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
	Content  string                           `json:"content" binding:"required"`
}

type BulkUpdatePipelineRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
	Content  string                           `json:"content" binding:"required"`
}

type BulkDeletePipelineRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
}

type BulkActivatePipelineRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
	Active   bool                             `json:"active"`
}
