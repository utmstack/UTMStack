package connectors

import (
	"context"

	"github.com/google/uuid"

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
}

type ExecutionUsecase interface {
	List(ctx context.Context, f dto.ExecutionFilters) (*database.List[dto.ExecutionResponse], error)
	HandleMatch(ctx context.Context, req dto.MatchRequest) error
	// StartFlow enqueues a flow run from raw alert JSON without alert matching.
	// The manual-trigger path; agentOverride wins over alert.dataSource for
	// shell roots. agentOverride is empty when the caller supplies no hint.
	StartFlow(ctx context.Context, relPath string, alertJSON []byte, agentOverride, triggeredBy string, manual bool) (uuid.UUID, error)
	StartManual(ctx context.Context, agent, command, triggeredBy string) (uuid.UUID, error)
	FinishManual(ctx context.Context, id uuid.UUID, status domain.ExecutionStatus, result string) error
}

type VariableUsecase interface {
	Create(ctx context.Context, req dto.CreateVariableRequest, user string) (*dto.VariableResponse, error)
	Update(ctx context.Context, req dto.UpdateVariableRequest, user string) (*dto.VariableResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*dto.VariableResponse, error)
	FindAll(ctx context.Context, f dto.VariableFilter) ([]dto.VariableResponse, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error

	InterpolateCommand(ctx context.Context, cmd, shell string) (string, error)
	MaskSecrets(ctx context.Context, output string) (string, error)
}

type AgentUsecase interface {
	ListByPlatform(ctx context.Context, platform string) ([]string, error)
}
