package dto

import "time"

// Endpoint is one agent row as the EDR console sees it, folded with its
// stored EDR status (when any).
type Endpoint struct {
	AgentID  uint32    `json:"agentId"`
	Hostname string    `json:"hostname"`
	IP       string    `json:"ip"`
	OS       string    `json:"os"`
	Platform string    `json:"platform"`
	Version  string    `json:"version"`
	Status   string    `json:"status"` // agent connectivity (online/offline)
	LastSeen time.Time `json:"lastSeen"`

	// EDR module state. HasModule is false when the endpoint never shipped a
	// status report — it has no EDR installed.
	HasModule      bool       `json:"hasModule"`
	ModuleHealthy  bool       `json:"moduleHealthy"`
	SigDBVersion   string     `json:"sigdbVersion,omitempty"`
	StaleSignature bool       `json:"staleSignature"`
	PolicyVersion  string     `json:"policyVersion,omitempty"`
	Drift          bool       `json:"drift"`
	LastReport     *time.Time `json:"lastReport,omitempty"`
}
