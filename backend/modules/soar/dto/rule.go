package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type FilterVM struct {
	Operator domain.OperatorType `json:"operator"`
	Field    string              `json:"field"`
	Value    any                 `json:"value"`
}

type FlowCommandVM struct {
	Command   string            `json:"command"             binding:"required"`
	Condition *domain.Condition `json:"condition,omitempty" binding:"omitempty,oneof=OnSuccess OnFailure Always"`
}

type CreateRuleRequest struct {
	ID             *int64          `json:"id"`
	Name           string          `json:"name"           binding:"required,max=150"`
	Description    string          `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM      `json:"conditions"     binding:"required,min=1"`
	Commands       []FlowCommandVM `json:"commands"       binding:"required,min=1,dive"`
	Active         *bool           `json:"active"         binding:"required"`
	AgentPlatform  string          `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string          `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string          `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string        `json:"excludedAgents"`
}

type UpdateRuleRequest struct {
	ID             *int64          `json:"id"`
	Name           string          `json:"name"           binding:"required,max=150"`
	Description    string          `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM      `json:"conditions"     binding:"required,min=1"`
	Commands       []FlowCommandVM `json:"commands"       binding:"required,min=1,dive"`
	Active         *bool           `json:"active"         binding:"required"`
	AgentPlatform  string          `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string          `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string          `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string        `json:"excludedAgents"`
}

type ToggleRuleRequest struct {
	Enabled bool `json:"enabled"`
}

type RuleResponse struct {
	RelPath          string          `json:"relPath"`
	Name             string          `json:"name"`
	Description      string          `json:"description,omitempty"`
	Conditions       []FilterVM      `json:"conditions"`
	Commands         []FlowCommandVM `json:"commands"`
	Active           bool            `json:"active"`
	AgentPlatform    string          `json:"agentPlatform,omitempty"`
	DefaultAgent     string          `json:"defaultAgent,omitempty"`
	Shell            string          `json:"shell,omitempty"`
	ExcludedAgents   []string        `json:"excludedAgents,omitempty"`
	SystemOwner      bool            `json:"systemOwner"`
	LastModifiedDate *time.Time      `json:"lastModifiedDate,omitempty"`
}

type RuleFilters struct {
	ID                  int64  `form:"id.equals"`
	RuleName            string `form:"name.contains"`
	RuleActive          *bool  `form:"active.equals"`
	AgentPlatform       string `form:"agentPlatform.equals"`
	CreatedBy           string `form:"createdBy.equals"`
	LastModifiedBy      string `form:"lastModifiedBy.equals"`
	CreatedDateGTE      string `form:"createdDate.greaterThanOrEqual"`
	CreatedDateLTE      string `form:"createdDate.lessThanOrEqual"`
	LastModifiedDateGTE string `form:"lastModifiedDate.greaterThanOrEqual"`
	LastModifiedDateLTE string `form:"lastModifiedDate.lessThanOrEqual"`
	SystemOwner         *bool  `form:"systemOwner.equals"`
	database.Params
}
