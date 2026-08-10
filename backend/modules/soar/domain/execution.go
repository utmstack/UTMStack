package domain

import (
	"time"

	"github.com/google/uuid"
)

type ExecutionStatus string

const (
	ExecutionStatusExecuted ExecutionStatus = "EXECUTED"
	ExecutionStatusPending  ExecutionStatus = "PENDING"
	ExecutionStatusFailed   ExecutionStatus = "FAILED"
)

type NonExecutionCause string

const (
	NonExecutionCauseAgentOffline  NonExecutionCause = "AGENT_OFFLINE"
	NonExecutionCauseAgentNotFound NonExecutionCause = "AGENT_NOT_FOUND"
	NonExecutionCauseUnknown       NonExecutionCause = "UNKNOWN"
)

type ExecutionOrigin string

const (
	ExecutionOriginFlow   ExecutionOrigin = "FLOW"
	ExecutionOriginManual ExecutionOrigin = "MANUAL"
)

type SoarExecution struct {
	ID                uuid.UUID          `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"                                 json:"id"`
	TenantID          uuid.UUID          `gorm:"column:tenant_id;type:uuid;not null;index:idx_soar_execution_tenant_started,priority:1"   json:"-"`
	Origin            ExecutionOrigin    `gorm:"column:origin;size:20;not null"      json:"origin"`
	RulePath          string             `gorm:"column:rule_path;size:512;not null"  json:"rulePath,omitempty"`      // flow only — the YAML file, not an FK
	AlertID           string             `gorm:"column:alert_id;size:150;not null"   json:"alertId,omitempty"`       // flow only
	TriggeredBy       string             `gorm:"column:triggered_by;size:150;not null" json:"triggeredBy,omitempty"` // manual only
	Agent             string             `gorm:"column:agent;size:150;not null" json:"agent"`
	Command           string             `gorm:"column:command;not null"        json:"command"`
	Result            string             `gorm:"column:result"                  json:"result,omitempty"`
	Status            ExecutionStatus    `gorm:"column:status;size:100;not null" json:"status"`
	StartedAt         time.Time          `gorm:"column:started_at;not null;index:idx_soar_execution_tenant_started,priority:2,sort:desc" json:"startedAt"`
	FinishedAt        *time.Time         `gorm:"column:finished_at"                                                                     json:"finishedAt,omitempty"`
	ClaimedAt         *time.Time         `gorm:"column:claimed_at"                     json:"-"`
	Retries           int                `gorm:"column:retries;not null;default:0"     json:"retries"`
	NonExecutionCause *NonExecutionCause `gorm:"column:non_execution_cause;size:100"   json:"nonExecutionCause,omitempty"`
}

func (SoarExecution) TableName() string { return "soar_executions" }
