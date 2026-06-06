package domain

import "time"

type UtmRegexPattern struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement"`
	PatternID          string     `gorm:"column:pattern_id;size:100;not null"`
	PatternDescription string     `gorm:"column:pattern_description"`
	PatternDefinition  string     `gorm:"column:pattern_definition;not null"`
	SystemOwner        bool       `gorm:"column:system_owner;not null"`
	LastUpdate         *time.Time `gorm:"column:last_update"`
}

func (UtmRegexPattern) TableName() string { return "utm_regex_pattern" }
