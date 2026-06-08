package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
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

// CreateExecutionRequest is the payload internal callers (the SOAR plugin) send
// to record a freshly-matched rule. Status is forced to PENDING server-side.
type CreateExecutionRequest struct {
	RuleID  int64  `json:"ruleId"  binding:"required"`
	AlertID string `json:"alertId" binding:"required,max=150"`
	Command string `json:"command" binding:"required"`
	Agent   string `json:"agent"   binding:"required,max=150"`
}

// UpdateExecutionRequest is a partial PATCH — only non-nil fields are written.
// IncrementRetries=true does a server-side `retries + 1` to avoid races between
// concurrent dispatcher ticks.
type UpdateExecutionRequest struct {
	ExecutionStatus   *domain.ExecutionStatus    `json:"executionStatus,omitempty"   binding:"omitempty,oneof=PENDING EXECUTED FAILED"`
	CommandResult     *string                    `json:"commandResult,omitempty"`
	NonExecutionCause *domain.NonExecutionCause  `json:"nonExecutionCause,omitempty" binding:"omitempty,oneof=AGENT_OFFLINE AGENT_NOT_FOUND UNKNOWN"`
	IncrementRetries  bool                       `json:"incrementRetries,omitempty"`
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
	database.Params
}
