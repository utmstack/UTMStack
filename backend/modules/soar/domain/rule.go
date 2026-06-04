package domain

import "time"

type AlertResponseRule struct {
	ID               int64      `gorm:"column:id;primaryKey"                              json:"id,omitempty"`
	RuleName         string     `gorm:"column:rule_name;size:150;not null"                json:"name"`
	RuleDescription  string     `gorm:"column:rule_description;size:512"                  json:"description,omitempty"`
	RuleConditions   string     `gorm:"column:rule_conditions;not null"                   json:"conditions"` // JSON string of []FilterType
	RuleCmd          string     `gorm:"column:rule_cmd;not null"                          json:"command"`
	RuleActive       bool       `gorm:"column:rule_active;not null"                       json:"active"`
	AgentPlatform    string     `gorm:"column:agent_platform"                             json:"agentPlatform,omitempty"`
	ExcludedAgents   string     `gorm:"column:excluded_agents"                            json:"excludedAgents,omitempty"` // CSV string
	DefaultAgent     string     `gorm:"column:default_agent;size:500"                     json:"defaultAgent,omitempty"`
	RuleShell        string     `gorm:"column:rule_shell;size:20"                         json:"shell,omitempty"`
	SystemOwner      bool       `gorm:"column:system_owner;not null"                      json:"systemOwner"`
	CreatedBy        string     `gorm:"column:created_by;size:50;not null;->:false;<-:create"  json:"createdBy,omitempty"`
	CreatedDate      *time.Time `gorm:"column:created_date;->:false;<-:create"            json:"createdDate,omitempty"`
	LastModifiedBy   string     `gorm:"column:last_modified_by;size:50"                   json:"lastModifiedBy,omitempty"`
	LastModifiedDate *time.Time `gorm:"column:last_modified_date"                         json:"lastModifiedDate,omitempty"`

	// Associations
	Executions []AlertResponseRuleExecution  `gorm:"foreignKey:RuleID"                          json:"executions,omitempty"`
	Templates  []AlertResponseActionTemplate `gorm:"many2many:utm_alert_response_rule_template;joinForeignKey:rule_id;joinReferences:template_id" json:"actions,omitempty"`
}

func (AlertResponseRule) TableName() string { return "utm_alert_response_rule" }
