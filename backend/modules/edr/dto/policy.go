package dto

import "time"

// PolicyRequest is the create/update body for a policy document.
type PolicyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Version     string `json:"version" binding:"required"`
	Doc         string `json:"doc" binding:"required"`
}

// PolicyResponse is the API shape for one policy.
type PolicyResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Doc         string    `json:"doc"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// AssignRequest targets endpoints for a policy.
type AssignRequest struct {
	AgentIDs []uint   `json:"agentIds" binding:"required"`
	Reason   string   `json:"reason"`
	Allow    []string `json:"allow,omitempty"` // optional allowlist entries to add
	Block    []string `json:"block,omitempty"` // optional blocklist entries to add
}
