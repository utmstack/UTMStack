package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionStatusWaiting   ExecutionStatus = "WAITING"
	ExecutionStatusPending   ExecutionStatus = "PENDING"
	ExecutionStatusExecuting ExecutionStatus = "EXECUTING"
	ExecutionStatusExecuted  ExecutionStatus = "EXECUTED"
	ExecutionStatusFailed    ExecutionStatus = "FAILED"
	ExecutionStatusDead      ExecutionStatus = "DEAD"
)

// Terminal returns true if the status represents a finished execution — the
// dispatcher will not touch it again.
func (s ExecutionStatus) Terminal() bool {
	switch s {
	case ExecutionStatusExecuted, ExecutionStatusFailed, ExecutionStatusDead:
		return true
	}
	return false
}

type NonExecutionCause string

const (
	NonExecutionCauseAgentOffline  NonExecutionCause = "AGENT_OFFLINE"
	NonExecutionCauseAgentNotFound NonExecutionCause = "AGENT_NOT_FOUND"
	NonExecutionCauseMaxDepth      NonExecutionCause = "MAX_DEPTH_EXCEEDED"
	NonExecutionCauseUnknown       NonExecutionCause = "UNKNOWN"
)

type ExecutionOrigin string

const (
	ExecutionOriginFlow   ExecutionOrigin = "FLOW"
	ExecutionOriginManual ExecutionOrigin = "MANUAL"
)

type EdgeBranch string

const (
	EdgeBranchSuccess EdgeBranch = "SUCCESS"
	EdgeBranchError   EdgeBranch = "ERROR"
)

// SoarExecution is one node instance of a running flow. A single flow run may
// hold many rows; siblings that AND-join land on the same (flow_run_id, node_id,
// depth) tuple and coalesce.
type SoarExecution struct {
	ID       uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"                               json:"id"`
	TenantID uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;index:idx_soar_execution_tenant_started,priority:1" json:"-"`
	Origin   ExecutionOrigin `gorm:"column:origin;size:20;not null" json:"origin"`

	// Manual-execution fields (untouched by the DAG engine).
	TriggeredBy string `gorm:"column:triggered_by;size:150;not null" json:"triggeredBy,omitempty"`

	// Flow-run linkage — nullable for manual executions.
	FlowRunID *uuid.UUID `gorm:"column:flow_run_id;type:uuid;index" json:"flowRunId,omitempty"`
	RulePath  string     `gorm:"column:rule_path;size:512;not null" json:"rulePath,omitempty"`
	AlertID   string     `gorm:"column:alert_id;size:150;not null"  json:"alertId,omitempty"`

	// DAG node identity.
	NodeID string   `gorm:"column:node_id;size:150;not null;index:idx_soar_execution_run_node,unique,priority:2" json:"nodeId,omitempty"`
	Depth  int      `gorm:"column:depth;not null;default:0;index:idx_soar_execution_run_node,unique,priority:3"  json:"depth"`
	Kind   NodeKind `gorm:"column:kind;size:20;not null"                                                          json:"kind"`

	// Executor + interpolated payload.
	Executor string          `gorm:"column:executor;size:60;not null"     json:"executor"`
	Params   json.RawMessage `gorm:"column:params;type:jsonb"             json:"params,omitempty"`
	Output   json.RawMessage `gorm:"column:output;type:jsonb"             json:"output,omitempty"`
	Context  json.RawMessage `gorm:"column:context;type:jsonb"            json:"context,omitempty"`

	// Shell-executor legacy fields (kept for manual execs and shell nodes).
	Agent   string `gorm:"column:agent;size:150;not null" json:"agent"`
	Command string `gorm:"column:command;not null"        json:"command"`
	Result  string `gorm:"column:result"                  json:"result,omitempty"`
	Shell   string `gorm:"column:shell;size:20"           json:"shell,omitempty"`

	// AND-join accounting.
	PendingParents int `gorm:"column:pending_parents;not null;default:0" json:"pendingParents"`
	DeadParents    int `gorm:"column:dead_parents;not null;default:0"    json:"deadParents"`

	Status            ExecutionStatus    `gorm:"column:status;size:20;not null" json:"status"`
	StartedAt         time.Time          `gorm:"column:started_at;not null;index:idx_soar_execution_tenant_started,priority:2,sort:desc" json:"startedAt"`
	FinishedAt        *time.Time         `gorm:"column:finished_at"                                                                      json:"finishedAt,omitempty"`
	ClaimedAt         *time.Time         `gorm:"column:claimed_at"                                                                       json:"-"`
	Retries           int                `gorm:"column:retries;not null;default:0"                                                       json:"retries"`
	NonExecutionCause *NonExecutionCause `gorm:"column:non_execution_cause;size:100"                                                     json:"nonExecutionCause,omitempty"`
}

func (SoarExecution) TableName() string { return "soar_executions" }

// SoarFlowRun groups every node execution triggered by a single alert match.
type SoarFlowRun struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID   uuid.UUID       `gorm:"column:tenant_id;type:uuid;not null;index"                json:"-"`
	RulePath   string          `gorm:"column:rule_path;size:512;not null"                       json:"rulePath"`
	AlertID    string          `gorm:"column:alert_id;size:150;not null"                        json:"alertId"`
	AlertJSON  json.RawMessage `gorm:"column:alert_json;type:jsonb;not null"                    json:"-"`
	MaxDepth   int             `gorm:"column:max_depth;not null;default:50"                     json:"maxDepth"`
	Status     ExecutionStatus `gorm:"column:status;size:20;not null"                           json:"status"`
	StartedAt  time.Time       `gorm:"column:started_at;not null"                               json:"startedAt"`
	FinishedAt *time.Time      `gorm:"column:finished_at"                                       json:"finishedAt,omitempty"`
}

func (SoarFlowRun) TableName() string { return "soar_flow_runs" }

// SoarExecutionEdge records a resolved incoming edge on a child execution.
// `Fired=false` means the parent branch didn't match — the edge died and the
// child is no longer reachable through this path.
type SoarExecutionEdge struct {
	ChildExecID  uuid.UUID  `gorm:"column:child_exec_id;type:uuid;primaryKey"  json:"childExecId"`
	ParentExecID uuid.UUID  `gorm:"column:parent_exec_id;type:uuid;primaryKey" json:"parentExecId"`
	FlowRunID    uuid.UUID  `gorm:"column:flow_run_id;type:uuid;not null;index" json:"flowRunId"`
	Branch       EdgeBranch `gorm:"column:branch;size:16;not null"             json:"branch"`
	Fired        bool       `gorm:"column:fired;not null"                      json:"fired"`
	CreatedAt    time.Time  `gorm:"column:created_at;not null"                 json:"createdAt"`
}

func (SoarExecutionEdge) TableName() string { return "soar_execution_edges" }
