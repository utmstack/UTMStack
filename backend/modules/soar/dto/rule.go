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

type CreateRuleRequest struct {
	ID             *int64     `json:"id"`
	Name           string     `json:"name"           binding:"required,max=150"`
	Description    string     `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM `json:"conditions"     binding:"required,min=1"`
	Commands       []string   `json:"commands"       binding:"required,min=1"`
	Active         *bool      `json:"active"         binding:"required"`
	AgentPlatform  string     `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string     `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string     `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string   `json:"excludedAgents"`
}

type UpdateRuleRequest struct {
	ID             *int64     `json:"id"`
	Name           string     `json:"name"           binding:"required,max=150"`
	Description    string     `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM `json:"conditions"     binding:"required,min=1"`
	Commands       []string   `json:"commands"       binding:"required,min=1"`
	Active         *bool      `json:"active"         binding:"required"`
	AgentPlatform  string     `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string     `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string     `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string   `json:"excludedAgents"`
}

type ToggleRuleRequest struct {
	Enabled bool `json:"enabled"`
}

type RuleResponse struct {
	RelPath          string     `json:"relPath"`
	Name             string     `json:"name"`
	Description      string     `json:"description,omitempty"`
	Conditions       []FilterVM `json:"conditions"`
	Commands         []string   `json:"commands"`
	Active           bool       `json:"active"`
	AgentPlatform    string     `json:"agentPlatform,omitempty"`
	DefaultAgent     string     `json:"defaultAgent,omitempty"`
	Shell            string     `json:"shell,omitempty"`
	ExcludedAgents   []string   `json:"excludedAgents,omitempty"`
	SystemOwner      bool       `json:"systemOwner"`
	LastModifiedDate *time.Time `json:"lastModifiedDate,omitempty"`
}

type RuleFilters struct {
	// id.equals — exact match on rule ID (JHipster: LongFilter)
	ID int64 `form:"id.equals"`
	// name.contains — substring match on rule name (JHipster: StringFilter)
	RuleName string `form:"name.contains"`
	// active.equals — filter by active flag (JHipster: BooleanFilter)
	RuleActive *bool `form:"active.equals"`
	// agentPlatform.equals — exact match on agent platform (JHipster: StringFilter)
	AgentPlatform string `form:"agentPlatform.equals"`
	// createdBy.equals — exact match on createdBy (JHipster: StringFilter)
	CreatedBy string `form:"createdBy.equals"`
	// lastModifiedBy.equals — exact match on lastModifiedBy (JHipster: StringFilter)
	LastModifiedBy string `form:"lastModifiedBy.equals"`
	// createdDate.greaterThanOrEqual — inclusive lower bound (JHipster: InstantFilter, RFC3339)
	CreatedDateGTE string `form:"createdDate.greaterThanOrEqual"`
	// createdDate.lessThanOrEqual — inclusive upper bound (JHipster: InstantFilter, RFC3339)
	CreatedDateLTE string `form:"createdDate.lessThanOrEqual"`
	// lastModifiedDate.greaterThanOrEqual — inclusive lower bound (JHipster: InstantFilter, RFC3339)
	LastModifiedDateGTE string `form:"lastModifiedDate.greaterThanOrEqual"`
	// lastModifiedDate.lessThanOrEqual — inclusive upper bound (JHipster: InstantFilter, RFC3339)
	LastModifiedDateLTE string `form:"lastModifiedDate.lessThanOrEqual"`
	// systemOwner.equals — filter by system ownership (JHipster: BooleanFilter)
	SystemOwner *bool `form:"systemOwner.equals"`
	database.Params
}
