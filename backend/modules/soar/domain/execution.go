package domain

import "time"

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

type AlertResponseRuleExecution struct {
	ID                int64              `gorm:"column:id;primaryKey;autoIncrement"         json:"id"`
	RuleID            int64              `gorm:"column:rule_id;not null"                    json:"ruleId"`
	AlertID           string             `gorm:"column:alert_id;size:150;not null"           json:"alertId"`
	Command           string             `gorm:"column:command;not null"                    json:"command"`
	CommandResult     string             `gorm:"column:command_result"                      json:"commandResult,omitempty"`
	Agent             string             `gorm:"column:agent;size:150;not null"             json:"agent"`
	ExecutionDate     time.Time          `gorm:"column:execution_date;not null;->:false;<-:create"  json:"executionDate"`
	ExecutionStatus   ExecutionStatus    `gorm:"column:execution_status;size:100;not null"  json:"executionStatus"`
	NonExecutionCause *NonExecutionCause `gorm:"column:non_execution_cause;size:100"       json:"nonExecutionCause,omitempty"`
	ExecutionRetries  int                `gorm:"column:execution_retries;default:0"         json:"executionRetries"`

	Rule *AlertResponseRule `gorm:"foreignKey:RuleID;references:ID;->:false;<-:false" json:"rule,omitempty"`
}

func (AlertResponseRuleExecution) TableName() string { return "utm_alert_response_rule_execution" }
