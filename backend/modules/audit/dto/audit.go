package dto

import (
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type LogResponse struct {
	ID           uint64          `json:"id"`
	Timestamp    time.Time       `json:"timestamp"`
	UserID       *uint64         `json:"user_id,omitempty"`
	UserLogin    string          `json:"user_login,omitempty"`
	IP           string          `json:"ip,omitempty"`
	UserAgent    string          `json:"user_agent,omitempty"`
	SessionID    *uint64         `json:"session_id,omitempty"`
	Action       string          `json:"action"`
	Status       string          `json:"status"`
	ErrorMessage string          `json:"error_message,omitempty"`
	ResourceType string          `json:"resource_type,omitempty"`
	ResourceID   string          `json:"resource_id,omitempty"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

func ToResponse(a domain.AuditLog) LogResponse {
	var meta json.RawMessage
	if len(a.Metadata) > 0 {
		meta = json.RawMessage(a.Metadata)
	}
	return LogResponse{
		ID:           a.ID,
		Timestamp:    a.Timestamp,
		UserID:       a.UserID,
		UserLogin:    a.UserLogin,
		IP:           a.IP,
		UserAgent:    a.UserAgent,
		SessionID:    a.SessionID,
		Action:       a.Action,
		Status:       a.Status,
		ErrorMessage: a.ErrorMessage,
		ResourceType: a.ResourceType,
		ResourceID:   a.ResourceID,
		Metadata:     meta,
	}
}

// ListQuery is the audit list request: just the shared paging/search/sort.
// Filtering is expressed through the search_query DSL (field.op.value), not
// typed fields — the repository validates fields against an allowlist.
type ListQuery = common_models.ListRequest

// ListResponse is the paginated audit response.
type ListResponse = common_models.ListResponse[LogResponse]
