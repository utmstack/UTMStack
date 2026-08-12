package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

type BulkCreateCorrelationRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Rule     CreateCorrelationRuleRequest     `json:"rule"`
}

type BulkUpdateCorrelationRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Rule     UpdateCorrelationRuleRequest     `json:"rule"`
}

type BulkDeleteCorrelationRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
}

type BulkActivateCorrelationRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath" binding:"required"`
	Active   bool                             `json:"active"`
}
