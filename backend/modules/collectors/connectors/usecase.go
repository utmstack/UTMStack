package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/collectors/dto"
)

type CollectorUsecase interface {
	ListAndSync(ctx context.Context, q dto.CollectorListQuery) (*dto.ListCollectorsResponse, error)
	Delete(ctx context.Context, id int64) error
}

type AgentUsecase interface {
	ListAgents(ctx context.Context, q dto.AgentListQuery) ([]dto.AgentResponse, int32, error)
	ListAgentsWithCommands(ctx context.Context) ([]dto.AgentResponse, int32, error)
	ListAgentCommands(ctx context.Context, searchQuery string, page, size int32) ([]dto.AgentCommandResponse, int64, error)
	GetAgentByHostname(ctx context.Context, hostname string) (*dto.AgentResponse, error)
	CanRunCommand(ctx context.Context, hostname string) (bool, error)
	UpdateAgentAttrs(ctx context.Context, req dto.UpdateAgentRequest) (*dto.AgentAuthResponse, error)
}
