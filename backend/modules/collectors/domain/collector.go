package domain

import "time"

type UtmCollector struct {
	ID           int64      `gorm:"column:id;primaryKey;autoIncrement:false"`
	Status       string     `gorm:"column:status;not null"`
	CollectorKey string     `gorm:"column:collector_key;size:255;not null"`
	IP           string     `gorm:"column:ip;size:45;not null"`
	Hostname     string     `gorm:"column:hostname;size:255;not null"`
	Version      string     `gorm:"column:version;size:50"`
	Module       string     `gorm:"column:module;not null"`
	LastSeen     *time.Time `gorm:"column:last_seen"`
	GroupID      *int64     `gorm:"column:group_id"`
	Active       bool       `gorm:"column:active;not null;default:true"`
}

func (UtmCollector) TableName() string { return "utm_collectors" }
