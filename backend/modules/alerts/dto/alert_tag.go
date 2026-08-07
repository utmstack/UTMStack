package dto

import (
	"encoding/json"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type CreateAlertTagRequest struct {
	TagName  string `json:"tagName"  binding:"required"`
	TagColor string `json:"tagColor"`
}

type UpdateAlertTagRequest struct {
	ID          uuid.UUID `json:"id"          binding:"required"`
	TagName     string    `json:"tagName"     binding:"required"`
	TagColor    string    `json:"tagColor"`
	SystemOwner bool      `json:"systemOwner"`
}

type AlertTagFilters struct {
	TagName     *string
	SystemOwner *bool
	Page        int
	Size        int
}

type AlertTagResponse struct {
	ID          uuid.UUID `json:"id"`
	TagName     string    `json:"tagName"`
	TagColor    string    `json:"tagColor"`
	SystemOwner bool      `json:"systemOwner"`
}

type CreateAlertTagRuleRequest struct {
	Name        string                     `json:"name"        binding:"required"`
	Description string                     `json:"description" binding:"required"`
	Conditions  []common_models.FilterType `json:"conditions"  binding:"required"`
	Tags        []AlertTagRuleTagRef       `json:"tags"        binding:"required"`
}

type UpdateAlertTagRuleRequest struct {
	ID          uuid.UUID                  `json:"id"          binding:"required"`
	Name        string                     `json:"name"        binding:"required"`
	Description string                     `json:"description" binding:"required"`
	Conditions  []common_models.FilterType `json:"conditions"  binding:"required"`
	Tags        []AlertTagRuleTagRef       `json:"tags"        binding:"required"`
}

type AlertTagRuleTagRef struct {
	ID          uuid.UUID `json:"id"`
	TagName     string    `json:"tagName"`
	TagColor    string    `json:"tagColor"`
	SystemOwner bool      `json:"systemOwner"`
}

type AlertTagRuleFilters struct {
	ID             *uuid.UUID
	Name           *string
	ConditionField *string
	ConditionValue *string
	RuleActive     *bool
	RuleDeleted    *bool
	TagIDs         []uuid.UUID
	Page           int
	Size           int
}

type AlertTagRuleResponse struct {
	ID          uuid.UUID                  `json:"id"`
	Name        string                     `json:"name"`
	Description string                     `json:"description"`
	Conditions  []common_models.FilterType `json:"conditions"`
	Tags        []AlertTagRuleTagRef       `json:"tags"`
	Active      bool                       `json:"active"`
	Deleted     bool                       `json:"deleted"`
}

type ActiveAlertTagRule struct {
	TenantID        uuid.UUID       `json:"tenantId"`
	ID              uuid.UUID       `json:"id"`
	Name            string          `json:"name"`
	Conditions      json.RawMessage `json:"conditions"`
	AppliedTagNames []string        `json:"appliedTagNames"`
}
