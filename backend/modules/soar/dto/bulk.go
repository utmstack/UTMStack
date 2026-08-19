package dto

import "github.com/utmstack/utmstack/backend/pkg/common_models"

// BulkCreateRuleRequest creates a rule in each selected tenant.
type BulkCreateRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	Rule     CreateRuleRequest                `json:"rule"`
}

// BulkUpdateRuleRequest updates a rule (by relPath) in each selected tenant.
type BulkUpdateRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath"`
	Rule     UpdateRuleRequest                `json:"rule"`
}

// BulkDeleteRuleRequest deletes a rule (by relPath) in each selected tenant.
type BulkDeleteRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath"`
}

// BulkEnableRuleRequest enables or disables a rule in each selected tenant.
type BulkEnableRuleRequest struct {
	Selector common_models.BulkTenantSelector `json:"selector"`
	RelPath  string                           `json:"relPath"`
	Enabled  bool                             `json:"enabled"`
}
