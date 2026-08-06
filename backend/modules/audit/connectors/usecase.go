package connectors

import (
	"context"
	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/audit/dto"
)

type Logger interface {
	Log(ctx context.Context, ev Event)
}

type Event struct {
	Action        string         // "auth.login", "user.update", ...
	Status        string         // domain.StatusSuccess | domain.StatusFailure
	ResourceType  string         // "user", "alert", "rule", ...
	ResourceID    string         // optional but recommended; string to allow non-numeric ids
	ErrorMessage  string         // when Status = failure
	Metadata      map[string]any // free-form context; serialized to jsonb
	EventType     domain.ApplicationEventType
	UserEmail     string
	UserID        *uuid.UUID
	IP            string
	UserAgent     string
	SessionID     *uuid.UUID
	SupportAccess string // set when a platform administrator acted inside another tenant
}

type Usecase interface {
	List(ctx context.Context, q dto.ListQuery) (*dto.ListResponse, error)
	Get(ctx context.Context, id uint64) (*dto.LogResponse, error)
}
