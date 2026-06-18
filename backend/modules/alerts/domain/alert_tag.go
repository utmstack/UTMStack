package domain

import "time"

type UtmAlertTag struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	TagName     string `gorm:"size:50;not null;uniqueIndex"`
	TagColor    string `gorm:"size:15"`
	SystemOwner bool   `gorm:"not null;default:false"`
}

func (UtmAlertTag) TableName() string { return "utm_alert_tag" }

type UtmAlertTagRule struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement"`
	RuleName         string    `gorm:"type:text;not null;uniqueIndex"`
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
