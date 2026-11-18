package dto

import "time"

// EdrStatus is one stored status row from the agent-manager.
type EdrStatus struct {
	AgentID       uint32
	StatusJSON    string
	PolicyVersion string
	ReportedAt    time.Time
}
