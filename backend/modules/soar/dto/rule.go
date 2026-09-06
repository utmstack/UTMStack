package dto

import (
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type FilterVM struct {
	Operator domain.OperatorType `json:"operator"`
	Field    string              `json:"field"`
	Value    any                 `json:"value"`
}

// FlowNodeVM mirrors domain.FlowNode across the API. YAML on-disk uses the
// same field names — only the DAG shape lives here.
type FlowNodeVM struct {
	Kind           domain.NodeKind `json:"kind"                     binding:"required,oneof=executor enrichment"`
	Executor       string          `json:"executor"                 binding:"required,max=60"`
	Command        string          `json:"command,omitempty"`
	Shell          string          `json:"shell,omitempty"          binding:"omitempty,max=20"`
	Platform       string          `json:"platform,omitempty"       binding:"omitempty,max=60"`
	Agent          string          `json:"agent,omitempty"          binding:"omitempty,max=150"`
	ExcludedAgents []string        `json:"excludedAgents,omitempty"`
	Params         json.RawMessage `json:"params,omitempty"`
	OnSuccess      []string        `json:"onSuccess,omitempty"`
	OnError        []string        `json:"onError,omitempty"`
}

type CreateRuleRequest struct {
	ID          *int64                `json:"id"`
	Name        string                `json:"name"        binding:"required,max=150"`
	Description string                `json:"description" binding:"omitempty,max=512"`
	Conditions  []FilterVM            `json:"conditions"  binding:"required,min=1"`
	Roots       []string              `json:"roots"       binding:"required,min=1"`
	Nodes       map[string]FlowNodeVM `json:"nodes"       binding:"required,min=1"`
	MaxDepth    int                   `json:"maxDepth"    binding:"omitempty,min=1,max=1000"`
	Active      *bool                 `json:"active"      binding:"required"`
}

type UpdateRuleRequest struct {
	ID          *int64                `json:"id"`
	Name        string                `json:"name"        binding:"required,max=150"`
	Description string                `json:"description" binding:"omitempty,max=512"`
	Conditions  []FilterVM            `json:"conditions"  binding:"required,min=1"`
	Roots       []string              `json:"roots"       binding:"required,min=1"`
	Nodes       map[string]FlowNodeVM `json:"nodes"       binding:"required,min=1"`
	MaxDepth    int                   `json:"maxDepth"    binding:"omitempty,min=1,max=1000"`
	Active      *bool                 `json:"active"      binding:"required"`
}

type ToggleRuleRequest struct {
	Enabled bool `json:"enabled"`
}

type RuleResponse struct {
	RelPath          string                `json:"relPath"`
	Name             string                `json:"name"`
	Description      string                `json:"description,omitempty"`
	Conditions       []FilterVM            `json:"conditions"`
	Roots            []string              `json:"roots"`
	Nodes            map[string]FlowNodeVM `json:"nodes"`
	MaxDepth         int                   `json:"maxDepth,omitempty"`
	Active           bool                  `json:"active"`
	SystemOwner      bool                  `json:"systemOwner"`
	LastModifiedDate *time.Time            `json:"lastModifiedDate,omitempty"`
}

type RuleFilters struct {
	ID                  int64  `form:"id.equals"`
	RuleName            string `form:"name.contains"`
	RuleActive          *bool  `form:"active.equals"`
	CreatedBy           string `form:"createdBy.equals"`
	LastModifiedBy      string `form:"lastModifiedBy.equals"`
	CreatedDateGTE      string `form:"createdDate.greaterThanOrEqual"`
	CreatedDateLTE      string `form:"createdDate.lessThanOrEqual"`
	LastModifiedDateGTE string `form:"lastModifiedDate.greaterThanOrEqual"`
	LastModifiedDateLTE string `form:"lastModifiedDate.lessThanOrEqual"`
	SystemOwner         *bool  `form:"systemOwner.equals"`
	database.Params
}
