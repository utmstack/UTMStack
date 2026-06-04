package dto

type AgentListQuery struct {
	PageNumber  int    `form:"pageNumber"`
	PageSize    int    `form:"pageSize"`
	SearchQuery string `form:"searchQuery"`
	SortBy      string `form:"sortBy"`
}

type AgentResponse struct {
	ID             uint32 `json:"id"`
	IP             string `json:"ip"`
	Hostname       string `json:"hostname"`
	OS             string `json:"os"`
	Status         string `json:"status"`
	Platform       string `json:"platform"`
	Version        string `json:"version"`
	AgentKey       string `json:"agentKey"` // always "SECRET"
	LastSeen       string `json:"lastSeen"`
	Mac            string `json:"mac"`
	OsMajorVersion string `json:"osMajorVersion"`
	OsMinorVersion string `json:"osMinorVersion"`
	Aliases        string `json:"aliases"`
	Addresses      string `json:"addresses"`
}

type AgentCommandResponse struct {
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
	AgentID       uint32         `json:"agentId"`
	Command       string         `json:"command"`
	CommandStatus string         `json:"commandStatus"`
	Result        string         `json:"result"`
	ExecutedBy    string         `json:"executedBy"`
	CmdID         string         `json:"cmdId"`
	Reason        string         `json:"reason"`
	OriginType    string         `json:"originType"`
	OriginID      string         `json:"originId"`
	Agent         *AgentResponse `json:"agent,omitempty"`
}

type UpdateAgentRequest struct {
	IP       string `json:"ip"       binding:"required"`
	Hostname string `json:"hostname" binding:"required"`
}

type AgentAuthResponse struct {
	ID  uint32 `json:"id"`
	Key string `json:"key"`
}
