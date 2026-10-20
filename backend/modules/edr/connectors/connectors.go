package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/edr/domain"
	"github.com/utmstack/utmstack/backend/modules/edr/dto"
)

// Repository backs the EDR policy store.
type Repository interface {
	Create(ctx context.Context, p *domain.Policy) error
	Update(ctx context.Context, p *domain.Policy) error
	Delete(ctx context.Context, id uint) error
	Get(ctx context.Context, id uint) (*domain.Policy, error)
	List(ctx context.Context) ([]domain.Policy, error)
}

// AgentGateway is the view of the agent-manager the EDR console needs.
type AgentGateway interface {
	// ListEndpoints returns every agent in the caller's tenant.
	ListEndpoints(ctx context.Context) ([]dto.Endpoint, error)
	// EdrStatuses returns stored EDR status rows for the tenant's agents.
	EdrStatuses(ctx context.Context) ([]dto.EdrStatus, error)
	// Dispatch pushes one EDR action to one endpoint. Fire-and-forget: the
	// result lands in the command history when the endpoint answers.
	Dispatch(ctx context.Context, agentID uint32, action, payloadJSON, executedBy, reason string) (string, error)
}

// DetectionCounter counts EDR events in the logs store over a window.
type DetectionCounter interface {
	CountInWindow(ctx context.Context, from, to time.Time) (int64, error)
}
