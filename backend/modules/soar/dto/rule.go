package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type FilterVM struct {
	Operator domain.OperatorType `json:"operator"`
	Field    string              `json:"field"`
	Value    any                 `json:"value"`
}

type TemplateDigest struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Command     string `json:"command"`
	SystemOwner bool   `json:"systemOwner"`
}

type CreateRuleRequest struct {
	ID             *int64           `json:"id"`
	Name           string           `json:"name"           binding:"required,max=150"`
	Description    string           `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM       `json:"conditions"     binding:"required,min=1"`
	Command        string           `json:"command"        binding:"required"`
	Active         *bool            `json:"active"         binding:"required"`
	AgentPlatform  string           `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string           `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string           `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string         `json:"excludedAgents"`
	Actions        []TemplateDigest `json:"actions"`
}

type UpdateRuleRequest struct {
	ID             *int64           `json:"id"             binding:"required"`
	Name           string           `json:"name"           binding:"required,max=150"`
	Description    string           `json:"description"    binding:"omitempty,max=512"`
	Conditions     []FilterVM       `json:"conditions"     binding:"required,min=1"`
	Command        string           `json:"command"        binding:"required"`
	Active         *bool            `json:"active"         binding:"required"`
	AgentPlatform  string           `json:"agentPlatform"  binding:"required"`
	DefaultAgent   string           `json:"defaultAgent"   binding:"omitempty,max=500"`
	Shell          string           `json:"shell"          binding:"omitempty,max=20"`
	ExcludedAgents []string         `json:"excludedAgents"`
	Actions        []TemplateDigest `json:"actions"`
}

type RuleResponse struct {
	ID               int64            `json:"id"`
	Name             string           `json:"name"`
	Description      string           `json:"description,omitempty"`
	Conditions       []FilterVM       `json:"conditions"`
	Command          string           `json:"command"`
	Active           bool             `json:"active"`
	AgentPlatform    string           `json:"agentPlatform,omitempty"`
	DefaultAgent     string           `json:"defaultAgent,omitempty"`
	Shell            string           `json:"shell,omitempty"`
	ExcludedAgents   []string         `json:"excludedAgents,omitempty"`
	SystemOwner      bool             `json:"systemOwner"`
	Actions          []TemplateDigest `json:"actions,omitempty"`
	CreatedBy        string           `json:"createdBy,omitempty"`
	CreatedDate      *time.Time       `json:"createdDate,omitempty"`
	LastModifiedBy   string           `json:"lastModifiedBy,omitempty"`
	LastModifiedDate *time.Time       `json:"lastModifiedDate,omitempty"`
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
	// Page is 0-based (matches Java Spring Pageable). Default 0.
	Page int `form:"page,default=0"`
	// Size is the page size (matches Java). Default 20.
	Size int `form:"size,default=20"`
}
