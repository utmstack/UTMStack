package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/collectors/connectors"
	"github.com/utmstack/utmstack/backend/modules/collectors/dto"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type agentUsecase struct {
	agentRepo agentRepo
}

type agentRepo interface {
	ListAgents(ctx context.Context, searchQuery string) ([]*agent.Agent, int32, error)
	ListAgentsPaged(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) ([]*agent.Agent, int32, error)
	ListAgentCommands(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListAgentsCommandsResponse, error)
	UpdateAgent(ctx context.Context, agentID uint32, agentKey, ip, hostname string) (*agent.AuthResponse, error)
}

func NewAgentUsecase(agentRepo agentRepo) connectors.AgentUsecase {
	return &agentUsecase{agentRepo: agentRepo}
}

func (u *agentUsecase) ListAgents(ctx context.Context, q dto.AgentListQuery) ([]dto.AgentResponse, int32, error) {
	pageNumber := int32(q.PageNumber)
	if pageNumber <= 0 {
		pageNumber = 1
	}
	pageSize := int32(q.PageSize)
	if pageSize <= 0 {
		pageSize = 1000000
	}

	rows, total, err := u.agentRepo.ListAgentsPaged(ctx, q.SearchQuery, pageNumber, pageSize, q.SortBy)
	if err != nil {
		return nil, 0, fmt.Errorf("agentUsecase.ListAgents: %w", err)
	}

	dtos := make([]dto.AgentResponse, 0, len(rows))
	for _, a := range rows {
		dtos = append(dtos, agentToDTO(a))
	}
	return dtos, total, nil
}

func (u *agentUsecase) ListAgentsWithCommands(ctx context.Context) ([]dto.AgentResponse, int32, error) {
	rows, total, err := u.agentRepo.ListAgentsPaged(ctx, "", 1, 1000000, "")
	if err != nil {
		return nil, 0, fmt.Errorf("agentUsecase.ListAgentsWithCommands: %w", err)
	}

	dtos := make([]dto.AgentResponse, 0, len(rows))
	for _, a := range rows {
		dtos = append(dtos, agentToDTO(a))
	}
	return dtos, total, nil
}

func (u *agentUsecase) ListAgentCommands(ctx context.Context, searchQuery string, page, size int32) ([]dto.AgentCommandResponse, int64, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}

	resp, err := u.agentRepo.ListAgentCommands(ctx, searchQuery, page, size, "")
	if err != nil {
		return nil, 0, fmt.Errorf("agentUsecase.ListAgentCommands: %w", err)
	}

	result := make([]dto.AgentCommandResponse, 0, len(resp.GetRows()))
	for _, cmd := range resp.GetRows() {
		r := agentCommandToDTO(cmd)

		// Enrich with nested agent (masked agentKey) — Java parity.
		agentRows, _, agentErr := u.agentRepo.ListAgents(ctx, fmt.Sprintf("id.Is=%d", cmd.GetAgentId()))
		if agentErr == nil && len(agentRows) > 0 {
			ag := agentToDTO(agentRows[0])
			r.Agent = &ag
		}

		result = append(result, r)
	}

	return result, int64(resp.GetTotal()), nil
}

func agentCommandToDTO(cmd *agent.AgentCommand) dto.AgentCommandResponse {
	return dto.AgentCommandResponse{
		CreatedAt:     formatTimestamp(cmd.GetCreatedAt()),
		UpdatedAt:     formatTimestamp(cmd.GetUpdatedAt()),
		AgentID:       cmd.GetAgentId(),
		Command:       cmd.GetCommand(),
		CommandStatus: cmd.GetCommandStatus().String(),
		Result:        cmd.GetResult(),
		ExecutedBy:    cmd.GetExecutedBy(),
		CmdID:         cmd.GetCmdId(),
		Reason:        cmd.GetReason(),
		OriginType:    cmd.GetOriginType(),
		OriginID:      cmd.GetOriginId(),
	}
}

func formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().UTC().Format("2006-01-02 15:04:05")
}

func (u *agentUsecase) GetAgentByHostname(ctx context.Context, hostname string) (*dto.AgentResponse, error) {
	rows, _, err := u.agentRepo.ListAgents(ctx, fmt.Sprintf("hostname.Is=%s", hostname))
	if err != nil {
		return nil, fmt.Errorf("agentUsecase.GetAgentByHostname: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("agent with hostname %q not found", hostname)
	}
	resp := agentToDTO(rows[0])
	return &resp, nil
}

func (u *agentUsecase) CanRunCommand(ctx context.Context, hostname string) (bool, error) {
	rows, _, err := u.agentRepo.ListAgents(ctx, fmt.Sprintf("hostname.Is=%s", hostname))
	if err != nil {
		return false, fmt.Errorf("agentUsecase.CanRunCommand: %w", err)
	}
	if len(rows) == 0 {
		return false, fmt.Errorf("agentUsecase.CanRunCommand: agent with hostname %q not found", hostname)
	}
	return rows[0].GetStatus() == agent.Status_ONLINE, nil
}

func (u *agentUsecase) UpdateAgentAttrs(ctx context.Context, req dto.UpdateAgentRequest) (*dto.AgentAuthResponse, error) {
	rows, _, err := u.agentRepo.ListAgents(ctx, fmt.Sprintf("hostname.Is=%s", req.Hostname))
	if err != nil {
		return nil, fmt.Errorf("agentUsecase.UpdateAgentAttrs resolve: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("agent with hostname %q not found", req.Hostname)
	}

	ag := rows[0]
	resp, err := u.agentRepo.UpdateAgent(ctx, ag.GetId(), ag.GetAgentKey(), req.IP, req.Hostname)
	if err != nil {
		return nil, fmt.Errorf("agentUsecase.UpdateAgentAttrs: %w", err)
	}

	return &dto.AgentAuthResponse{
		ID:  resp.GetId(),
		Key: resp.GetKey(),
	}, nil
}

func agentToDTO(a *agent.Agent) dto.AgentResponse {
	return dto.AgentResponse{
		ID:             a.GetId(),
		IP:             a.GetIp(),
		Hostname:       a.GetHostname(),
		OS:             a.GetOs(),
		Status:         a.GetStatus().String(),
		Platform:       a.GetPlatform(),
		Version:        a.GetVersion(),
		AgentKey:       "SECRET",
		LastSeen:       a.GetLastSeen(),
		Mac:            a.GetMac(),
		OsMajorVersion: a.GetOsMajorVersion(),
		OsMinorVersion: a.GetOsMinorVersion(),
		Aliases:        a.GetAliases(),
		Addresses:      a.GetAddresses(),
	}
}
