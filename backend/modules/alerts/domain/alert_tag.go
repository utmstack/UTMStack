package domain

import (
	"github.com/google/uuid"
)

type AlertTag struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_alert_tag_tenant_name" json:"-"`
	TagName     string    `gorm:"size:50;not null;uniqueIndex:idx_alert_tag_tenant_name"`
	TagColor    string    `gorm:"size:15"`
	SystemOwner bool      `gorm:"not null;default:false"`
}

func (AlertTag) TableName() string { return "alert_tag" }

func (AlertTag) SystemFlagColumn() string { return "system_owner" }

type AlertTagRule struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID        uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:idx_alert_tag_rule_tenant_name" json:"-"`
	RuleName        string    `gorm:"type:text;not null;uniqueIndex:idx_alert_tag_rule_tenant_name"`
	RuleDescription string    `gorm:"type:text;not null"`
	RuleConditions  string    `gorm:"type:text;not null"` // JSON-serialized []FilterType
	RuleAppliedTags string    `gorm:"type:text;not null"` // CSV of tag IDs — legacy smell
	RuleActive      bool      `gorm:"not null;default:false"`
	RuleDeleted     bool      `gorm:"not null;default:false"`
}

func (AlertTagRule) TableName() string { return "alert_tag_rule" }
