package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ExecutionResponse struct {
	ID                int64                     `json:"id"`
	RulePath          string                    `json:"rulePath"`
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
// One request per command — multi-command flows produce one row per command.
type CreateExecutionRequest struct {
	RulePath string `json:"rulePath" binding:"required,max=512"`
	AlertID  string `json:"alertId"  binding:"required,max=150"`
	Command  string `json:"command"  binding:"required"`
	Agent    string `json:"agent"    binding:"required,max=150"`
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
	// rulePath.equals — exact match on the flow's RelPath (JHipster: StringFilter)
	RulePath string `form:"rulePath.equals"`
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
