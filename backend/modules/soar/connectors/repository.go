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
	Get(ctx context.Context, id uuid.UUID) (*domain.SoarExecution, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, u ExecutionStatusUpdate) error
	ClaimPending(ctx context.Context, id uuid.UUID, leaseDuration time.Duration) (bool, error)

	// SaveOutput persists an enrichment node's structured output; called after
	// a successful Executor.Execute for kind=enrichment.
	SaveOutput(ctx context.Context, id uuid.UUID, output []byte) error

	// RecordEdge inserts one parent→child edge and atomically decrements the
	// child's pending_parents counter. If the child does not yet exist for
	// (flowRunID, childNodeID, childDepth), it is created with the given
	// template (status=WAITING, pending_parents=incomingCount, kind/executor
	// copied from the flow node). Returns the child's post-update state.
	RecordEdge(ctx context.Context, req RecordEdgeRequest) (child *domain.SoarExecution, err error)

	// ListFiredParents returns all fired parents of a child execution, used to
	// build the merged context bag when the child transitions to PENDING.
	ListFiredParents(ctx context.Context, childID uuid.UUID) ([]domain.SoarExecution, error)

	// TransitionReady moves a child from WAITING to PENDING (or DEAD) once all
	// its parents have resolved. context/params/command/shell are the
	// interpolated values computed by the caller from ListFiredParents.
	TransitionReady(ctx context.Context, id uuid.UUID, ready ReadyUpdate) error
}

type RecordEdgeRequest struct {
	FlowRunID       uuid.UUID
	TenantID        uuid.UUID
	RulePath        string
	AlertID         string
	Parent          domain.SoarExecution
	ChildNodeID     string
	ChildDepth      int
	ChildKind       domain.NodeKind
	ChildExecutor   string
	IncomingCount   int
	Branch          domain.EdgeBranch
	Fired           bool
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
