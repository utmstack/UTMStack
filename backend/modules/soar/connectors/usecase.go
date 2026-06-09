package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

type RuleUsecase interface {
	Create(ctx context.Context, req dto.CreateRuleRequest, createdBy string) (*dto.RuleResponse, error)
	Update(ctx context.Context, relPath string, req dto.UpdateRuleRequest, modifiedBy string) (*dto.RuleResponse, error)
	Get(ctx context.Context, relPath string) (*dto.RuleResponse, error)
	Delete(ctx context.Context, relPath string) error
	SetEnabled(ctx context.Context, relPath string, enabled bool) error
	List(ctx context.Context, f dto.RuleFilters) (*database.List[dto.RuleResponse], error)
	ResolveFilterValues(ctx context.Context) (*dto.ResolveFilterValuesResponse, error)
}

type TemplateUsecase interface {
	List(ctx context.Context, f dto.TemplateFilters) (*database.List[dto.TemplateResponse], error)
}

type ExecutionUsecase interface {
	List(ctx context.Context, f dto.ExecutionFilters) (*database.List[dto.ExecutionResponse], error)
	HandleMatch(ctx context.Context, req dto.MatchRequest) error
}

type VariableUsecase interface {
	Create(req dto.CreateVariableRequest, user string) (*dto.VariableResponse, error)
	Update(req dto.UpdateVariableRequest, user string) (*dto.VariableResponse, error)
	FindByID(id int64) (*dto.VariableResponse, error)
	FindAll(f dto.VariableFilter) ([]dto.VariableResponse, int64, error)
	Delete(id int64) error

	InterpolateCommand(cmd string) (string, error)
	MaskSecrets(output string) (string, error)
}

type ActionUsecase interface {
	Create(req dto.CreateActionRequest, user string) (*domain.UtmIncidentAction, error)
	Update(req dto.UpdateActionRequest, user string) (*domain.UtmIncidentAction, error)
	FindByID(id int64) (*domain.UtmIncidentAction, error)
	FindAll(f dto.ActionFilter) ([]domain.UtmIncidentAction, int64, error)
	Delete(id int64) error
}

type ActionCommandUsecase interface {
	Create(req dto.CreateActionCommandRequest) (*domain.UtmIncidentActionCommand, error)
	Update(req dto.UpdateActionCommandRequest) (*domain.UtmIncidentActionCommand, error)
	FindByID(id int64) (*domain.UtmIncidentActionCommand, error)
	FindAll(f dto.ActionCommandFilter) ([]domain.UtmIncidentActionCommand, int64, error)
	Delete(id int64) error
}

type AgentUsecase interface {
	ListByPlatform(ctx context.Context, platform string) ([]string, error)
}

type JobUsecase interface {
	Create(req dto.CreateJobRequest, user string) (*domain.UtmIncidentJob, error)
	FindByID(id int64) (*domain.UtmIncidentJob, error)
	FindAll(f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error)
	Count(f dto.JobFilter) (int64, error)
	Delete(id int64) error
}
