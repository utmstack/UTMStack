package dto

// QuarantineItem is one quarantined file on one endpoint.
type QuarantineItem struct {
	AgentID       uint32 `json:"agentId"`
	Hostname      string `json:"hostname"`
	ID            string `json:"id"`
	OriginalPath  string `json:"originalPath"`
	SHA256        string `json:"sha256"`
	Detection     string `json:"detection"`
	QuarantinedAt string `json:"quarantinedAt"`
	Restorable    bool   `json:"restorable"`
}
