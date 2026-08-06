package domain

import "time"

type UtmAlertTag struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	TenantID    string `gorm:"size:36;index;uniqueIndex:idx_alert_tag_tenant_name" json:"-"`
	TagName     string `gorm:"size:50;not null;uniqueIndex:idx_alert_tag_tenant_name"`
	TagColor    string `gorm:"size:15"`
	SystemOwner bool   `gorm:"not null;default:false"`
}

func (UtmAlertTag) TableName() string { return "utm_alert_tag" }

func (UtmAlertTag) SystemFlagColumn() string { return "system_owner" }

type UtmAlertTagRule struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	TenantID         string    `gorm:"size:36;index;uniqueIndex:idx_alert_tag_rule_tenant_name" json:"-"`
	RuleName         string    `gorm:"type:text;not null;uniqueIndex:idx_alert_tag_rule_tenant_name"`
	RuleDescription  string    `gorm:"type:text;not null"`
	RuleConditions   string    `gorm:"type:text;not null"` // JSON-serialized []FilterType
	RuleAppliedTags  string    `gorm:"type:text;not null"` // CSV of tag IDs — legacy smell
	RuleActive       bool      `gorm:"not null;default:false"`
	RuleDeleted      bool      `gorm:"not null;default:false"`
	CreatedBy        string    `gorm:"size:50;not null"`
	CreatedDate      time.Time `gorm:"not null;autoCreateTime"`
	LastModifiedBy   string    `gorm:"size:50"`
	LastModifiedDate *time.Time
}

func (UtmAlertTagRule) TableName() string { return "utm_alert_tag_rule" }
