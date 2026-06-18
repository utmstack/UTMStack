package domain

type RuleTemplate struct {
	RuleID     int64 `gorm:"column:rule_id;primaryKey"     json:"rule_id"`
	TemplateID int64 `gorm:"column:template_id;primaryKey" json:"template_id"`
}

func (RuleTemplate) TableName() string { return "utm_alert_response_rule_template" }
