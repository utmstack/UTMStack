package dto

import (
	"encoding/json"
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

type MatchRequest struct {
	RulePath string          `json:"rulePath" binding:"required,max=512"`
	Alert    json.RawMessage `json:"alert"    binding:"required"`
}

type ExecutionFilters struct {
	// id.equals — exact match on execution ID (JHipster: LongFilter)
	ID int64 `form:"id.equals"`
	// rulePath.equals — filter by the flow's file identity (relative path)
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
