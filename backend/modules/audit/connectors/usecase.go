package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/modules/audit/dto"
)

type Logger interface {
	Log(ctx context.Context, ev Event)
}

type Event struct {
	Action       string         // "auth.login", "user.update", ...
	Status       string         // domain.StatusSuccess | domain.StatusFailure
	ResourceType string         // "user", "alert", "rule", ...
	ResourceID   string         // optional but recommended; string to allow non-numeric ids
	ErrorMessage string         // when Status = failure
	Metadata     map[string]any // free-form context; serialized to jsonb
	EventType    domain.ApplicationEventType
	UserLogin    string
	UserID       *uint64
	IP           string
	UserAgent    string
	SessionID    *uint64
}

type Usecase interface {
	List(ctx context.Context, q dto.ListQuery) (*dto.ListResponse, error)
	Get(ctx context.Context, id uint64) (*dto.LogResponse, error)
}
