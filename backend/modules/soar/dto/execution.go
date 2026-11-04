package dto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ExecutionResponse struct {
	ID                uuid.UUID                 `json:"id"`
	Origin            domain.ExecutionOrigin    `json:"origin"`
	RulePath          string                    `json:"rulePath,omitempty"`
	AlertID           string                    `json:"alertId,omitempty"`
	TriggeredBy       string                    `json:"triggeredBy,omitempty"`
	Agent             string                    `json:"agent"`
	Command           string                    `json:"command"`
	Result            string                    `json:"result,omitempty"`
	Status            domain.ExecutionStatus    `json:"status"`
	StartedAt         time.Time                 `json:"startedAt"`
	FinishedAt        *time.Time                `json:"finishedAt,omitempty"`
	NonExecutionCause *domain.NonExecutionCause `json:"nonExecutionCause,omitempty"`
	Retries           int                       `json:"retries"`

	// DAG node tracking — populated for flow executions, empty for manual.
	NodeID    string          `json:"nodeId,omitempty"`
	Kind      domain.NodeKind `json:"kind,omitempty"`
	Executor  string          `json:"executor,omitempty"`
	FlowRunID *uuid.UUID      `json:"flowRunId,omitempty"`
	Depth     int             `json:"depth,omitempty"`
}

type MatchRequest struct {
	RulePath string          `json:"rulePath" binding:"required,max=512"`
	Alert    json.RawMessage `json:"alert"    binding:"required"`
}

// TriggerFlowRequest is the manual flow trigger. Alert is raw alert JSON (a
// past alert's stored payload, or a minimal synthetic one) — it feeds the
// $(alert.*) interpolation bag and agent auto-resolution. Agent optionally
// overrides dataSource for shell roots. Flow-level Conditions are ignored:
// a manual trigger means "run this flow now".
type TriggerFlowRequest struct {
	Alert json.RawMessage `json:"alert"`
	Agent string          `json:"agent" binding:"max=150"`
}

// FlowTriggerResponse reports the run id so the caller can follow it.
type FlowTriggerResponse struct {
	FlowRunID uuid.UUID `json:"flowRunId"`
}

type ExecutionFilters struct {
	Origin            domain.ExecutionOrigin   `form:"origin"`
	RulePath          string                   `form:"rulePath"`
	AlertID           string                   `form:"alertId"`
	Agent             string                   `form:"agent"`
	TriggeredBy       string                   `form:"triggeredBy"`
	Status            domain.ExecutionStatus   `form:"status"`
	NonExecutionCause domain.NonExecutionCause `form:"nonExecutionCause"`
	StartedAtGTE      string                   `form:"startedAtFrom"`
	StartedAtLTE      string                   `form:"startedAtTo"`
	database.Params
}
