package dto

// ActionResult is the outcome of one dispatched EDR action.
type ActionResult struct {
	AgentID  uint32 `json:"agentId"`
	Action   string `json:"action"`
	CmdID    string `json:"cmdId"`
	Dispatch bool   `json:"dispatched"`
	Error    string `json:"error,omitempty"`
}
