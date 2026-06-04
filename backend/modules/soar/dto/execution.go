package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

type ExecutionResponse struct {
	ID                int64                     `json:"id"`
	RuleID            int64                     `json:"ruleId"`
	AlertID           string                    `json:"alertId"`
	Command           string                    `json:"command"`
	CommandResult     string                    `json:"commandResult,omitempty"`
	Agent             string                    `json:"agent"`
	ExecutionDate     time.Time                 `json:"executionDate"`
	ExecutionStatus   domain.ExecutionStatus    `json:"executionStatus"`
	NonExecutionCause *domain.NonExecutionCause `json:"nonExecutionCause,omitempty"`
	ExecutionRetries  int                       `json:"executionRetries"`
}

type ExecutionFilters struct {
	// id.equals — exact match on execution ID (JHipster: LongFilter)
	ID int64 `form:"id.equals"`
	// ruleId.equals — filter by rule ID (JHipster: LongFilter)
	RuleID int64 `form:"ruleId.equals"`
	// ruleId.greaterThanOrEqual — inclusive lower bound on rule ID (JHipster: LongFilter)
	RuleIDGreaterThanOrEqual *int64 `form:"ruleId.greaterThanOrEqual"`
	// ruleId.lessThanOrEqual — inclusive upper bound on rule ID (JHipster: LongFilter)
	RuleIDLessThanOrEqual *int64 `form:"ruleId.lessThanOrEqual"`
	// alertId.contains — substring match on alertId (JHipster: StringFilter)
	AlertID string `form:"alertId.contains"`
	// agent.contains — substring match on agent name (JHipster: StringFilter)
	Agent string `form:"agent.contains"`
	// executionStatus.equals — exact match on status enum (JHipster: Filter<RuleExecutionStatus>)
	ExecutionStatus domain.ExecutionStatus `form:"executionStatus.equals"`
	// nonExecutionCause.equals — exact match on non-execution cause enum (JHipster: Filter<RuleNonExecutionCause>)
	NonExecutionCause domain.NonExecutionCause `form:"nonExecutionCause.equals"`
	// executionDate.greaterThanOrEqual — inclusive lower bound (JHipster: InstantFilter, RFC3339)
	ExecutionDateGTE string `form:"executionDate.greaterThanOrEqual"`
	// executionDate.lessThanOrEqual — inclusive upper bound (JHipster: InstantFilter, RFC3339)
	ExecutionDateLTE string `form:"executionDate.lessThanOrEqual"`
	// Page is 0-based (matches Java Spring Pageable). Default 0.
	Page int `form:"page,default=0"`
	// Size is the page size (matches Java). Default 20.
	Size int `form:"size,default=20"`
}
