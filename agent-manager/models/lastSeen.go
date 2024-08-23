package models

import "time"

type LastSeen struct {
	ConnectorType string
	ID            uint      `gorm:"primaryKey"`
	LastPing      time.Time `gorm:"index"`
}
