package models

import (
	"time"

	"gorm.io/gorm"
)

// AgentEdrStatus is the latest EDR status report stored per agent in the DB.
// The agent ships it on the AgentStream (every 5 min and on change); the
// console reads it without waking the endpoint. It survives an agent-manager
// restart and is never deleted by this path: an offline agent's last report
// stays readable. PolicyVersion carries the version of the last
// centrally-assigned policy as seen on the endpoint — the basis for drift
// detection.
type AgentEdrStatus struct {
	gorm.Model
	AgentID       uint   `gorm:"uniqueIndex;not null"`
	StatusJSON    string `gorm:"type:text"`
	PolicyVersion string
	ReportedAt    time.Time
}
