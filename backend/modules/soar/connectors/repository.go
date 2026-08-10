package connectors

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type ExecutionFilters struct {
	Origin            domain.ExecutionOrigin
	RulePath          string
	AlertID           string
	Agent             string
	TriggeredBy       string
	Status            domain.ExecutionStatus
	NonExecutionCause domain.NonExecutionCause
	StartedAtGTE      string
	StartedAtLTE      string
	database.Params
}

type ExecutionStatusUpdate struct {
	Status            *domain.ExecutionStatus
	Result            *string
	NonExecutionCause *domain.NonExecutionCause
	FinishedAt        *time.Time
	IncrementRetries  bool
}

type ExecutionRepository interface {
	Create(ctx context.Context, e *domain.SoarExecution) (*domain.SoarExecution, error)
	List(ctx context.Context, f ExecutionFilters) ([]domain.SoarExecution, int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, u ExecutionStatusUpdate) error
	ClaimPending(ctx context.Context, id uuid.UUID, leaseDuration time.Duration) (bool, error)
}

type ResolveFilterRepository interface {
	GetAgentPlatforms(ctx context.Context) ([]string, error)
	GetUsers(ctx context.Context) ([]string, error)
}

type AgentRepository interface {
	ListNamesByPlatform(ctx context.Context, platform string) ([]string, error)
}

type VariableRepository interface {
	Save(ctx context.Context, v *domain.SoarVariable) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.SoarVariable, error)
	FindAll(ctx context.Context, f dto.VariableFilter) ([]domain.SoarVariable, int64, error)
	FindAllPlain(ctx context.Context) ([]domain.SoarVariable, error)
	FindByName(ctx context.Context, name string) (*domain.SoarVariable, error)
	FindByNames(ctx context.Context, names []string) ([]domain.SoarVariable, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
