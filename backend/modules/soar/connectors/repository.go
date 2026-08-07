package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type TemplateFilters struct {
	ID          int64
	Label       string
	Description string
	Command     string
	SystemOwner *bool
	database.Params
}

type ExecutionFilters struct {
	ID                int64
	RulePath          string
	AlertID           string
	Agent             string
	ExecutionStatus   domain.ExecutionStatus
	NonExecutionCause domain.NonExecutionCause
	ExecutionDateGTE  string
	ExecutionDateLTE  string
	database.Params
}

type TemplateRepository interface {
	List(ctx context.Context, f TemplateFilters) ([]domain.AlertResponseActionTemplate, int64, error)
}

type ExecutionStatusUpdate struct {
	ExecutionStatus   *domain.ExecutionStatus
	CommandResult     *string
	NonExecutionCause *domain.NonExecutionCause
	IncrementRetries  bool
}

type ExecutionRepository interface {
	Create(ctx context.Context, e *domain.AlertResponseRuleExecution) (*domain.AlertResponseRuleExecution, error)
	List(ctx context.Context, f ExecutionFilters) ([]domain.AlertResponseRuleExecution, int64, error)
	UpdateStatus(ctx context.Context, id int64, u ExecutionStatusUpdate) error
	ClaimPending(ctx context.Context, id int64, leaseDuration time.Duration) (bool, error)
}

type ResolveFilterRepository interface {
	GetAgentPlatforms(ctx context.Context) ([]string, error)
	GetUsers(ctx context.Context) ([]string, error)
}

type AgentRepository interface {
	ListNamesByPlatform(ctx context.Context, platform string) ([]string, error)
}

type VariableRepository interface {
	Save(ctx context.Context, v *domain.UtmIncidentVariable) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentVariable, error)
	FindAll(ctx context.Context, f dto.VariableFilter) ([]domain.UtmIncidentVariable, int64, error)
	FindAllPlain(ctx context.Context) ([]domain.UtmIncidentVariable, error)
	FindByName(ctx context.Context, name string) (*domain.UtmIncidentVariable, error)
	FindByNames(ctx context.Context, names []string) ([]domain.UtmIncidentVariable, error)
	Delete(ctx context.Context, id int64) error
}

type ActionRepository interface {
	Save(ctx context.Context, action *domain.UtmIncidentAction) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentAction, error)
	FindAll(ctx context.Context, f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error)
	Delete(ctx context.Context, id int64) error
}

type ActionCommandRepository interface {
	Save(ctx context.Context, cmd *domain.UtmIncidentActionCommand) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentActionCommand, error)
	FindAll(ctx context.Context, f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error)
	Delete(ctx context.Context, id int64) error
}

type JobRepository interface {
	Save(ctx context.Context, job *domain.UtmIncidentJob) error
	FindByID(ctx context.Context, id int64) (*domain.UtmIncidentJob, error)
	FindAll(ctx context.Context, f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error)
	Count(ctx context.Context, f dto.JobFilter) (int64, error)
	Delete(ctx context.Context, id int64) error
}
