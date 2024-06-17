package models

import "time"

type Status int32

const (
	Online  Status = 0
	Offline Status = 1
	Unknown Status = 2
)

type LogCollectorClass string

const (
	CollectorClass LogCollectorClass = "collector"
	AgentClass     LogCollectorClass = "agent"
)

type LastSeen struct {
	Key      string    `gorm:"primaryKey"`
	LastPing time.Time `gorm:"index"`
}
