package dto

import "time"

// CommandRow is one edr command from the history.
type CommandRow struct {
	CmdID      string    `json:"cmdId"`
	Action     string    `json:"action"`
	Status     string    `json:"status"`
	Result     string    `json:"result,omitempty"`
	ExecutedBy string    `json:"executedBy"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}
