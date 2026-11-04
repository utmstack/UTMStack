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
	DispatchedAt      *time.Time
	IncrementRetries  bool
}

type ExecutionRepository interface {
	Create(ctx context.Context, e *domain.SoarExecution) (*domain.SoarExecution, error)
	List(ctx context.Context, f ExecutionFilters) ([]domain.SoarExecution, int64, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.SoarExecution, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, u ExecutionStatusUpdate) error
	// SettleExecuting is the worker's terminal write for a row it claimed:
	// it applies only while the row is still EXECUTING. It reports (false,
	// nil) when the status guard matches nothing — i.e. the reaper already
	// closed the row — so the caller must skip the child cascade: that row's
	// flow was already accounted by whoever won the race.
	SettleExecuting(ctx context.Context, id uuid.UUID, u ExecutionStatusUpdate) (bool, error)
	ClaimPending(ctx context.Context, id uuid.UUID, leaseDuration time.Duration) (bool, error)

	// SaveOutput persists an enrichment node's structured output; called after
	// a successful Executor.Execute for kind=enrichment.
	SaveOutput(ctx context.Context, id uuid.UUID, output []byte) error

	RecordEdge(ctx context.Context, req RecordEdgeRequest) (child *domain.SoarExecution, err error)

	RecordDeadBranch(ctx context.Context, req RecordEdgeRequest) error

	// RequeueStuck is the reaper's path for retriable executions: back to
	// PENDING with the claim cleared so the next tick picks the row up again.
	// Guarded by status: if the original worker settles the row between the
	// stuck list and this write, nothing happens and that path owns the row.
	RequeueStuck(ctx context.Context, id uuid.UUID) error

	// CloseStuckLost marks a stuck EXECUTING execution FAILED with the
	// dispatcher-lost cause. Like RequeueStuck it is guarded by status: it
	// reports (false, nil) — touching nothing — when the original worker
	// settled the row first, so the reaper must not also spawn its children.
	CloseStuckLost(ctx context.Context, id uuid.UUID) (bool, error)

	// ListStuckExecuting returns executions left EXECUTING past the claim lease
	// (their dispatcher worker died). Must run on an all-tenants context.
	ListStuckExecuting(ctx context.Context, cutoff time.Time) ([]domain.SoarExecution, error)

	ListFiredParents(ctx context.Context, childID uuid.UUID) ([]domain.SoarExecution, error)

	TransitionReady(ctx context.Context, id uuid.UUID, ready ReadyUpdate) error
}

type RecordEdgeRequest struct {
	FlowRunID     uuid.UUID
	TenantID      uuid.UUID
	RulePath      string
	AlertID       string
	Parent        domain.SoarExecution
	ChildNodeID   string
	ChildDepth    int
	ChildKind     domain.NodeKind
	ChildExecutor string
	Branch        domain.EdgeBranch
	Fired         bool
}

type ReadyUpdate struct {
	Status  domain.ExecutionStatus
	Context []byte
	Params  []byte
	Command string
	Shell   string
	Agent   string
}

// FlowRunRepository holds the top-level state of one root-invocation.
type FlowRunRepository interface {
	Create(ctx context.Context, r *domain.SoarFlowRun) (*domain.SoarFlowRun, error)
	Get(ctx context.Context, id uuid.UUID) (*domain.SoarFlowRun, error)
	// MaybeComplete transitions the run to COMPLETED/FAILED when no non-terminal
	// executions remain. Returns true when a transition happened.
	MaybeComplete(ctx context.Context, id uuid.UUID) (bool, error)
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
