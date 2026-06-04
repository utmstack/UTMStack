package repository

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
)

type agentManagerRepository struct {
	client *agentmanager.AgentManagerClient
}

func NewAgentManagerRepository(client *agentmanager.AgentManagerClient) *agentManagerRepository {
	return &agentManagerRepository{client: client}
}

func (r *agentManagerRepository) ListCollectors(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListCollectorResponse, error) {
	resp, err := r.client.ListCollectors(ctx, searchQuery, pageNumber, pageSize, sortBy)
	if err != nil {
		return nil, fmt.Errorf("agentManagerRepository.ListCollectors: %w", err)
	}
	return resp, nil
}

func (r *agentManagerRepository) ListAgents(ctx context.Context, searchQuery string) ([]*agent.Agent, int32, error) {
	rows, total, err := r.client.ListAgents(ctx, searchQuery)
	if err != nil {
		return nil, 0, fmt.Errorf("agentManagerRepository.ListAgents: %w", err)
	}
	return rows, total, nil
}

func (r *agentManagerRepository) ListAgentsPaged(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) ([]*agent.Agent, int32, error) {
	rows, total, err := r.client.ListAgentsPaged(ctx, searchQuery, pageNumber, pageSize, sortBy)
	if err != nil {
		return nil, 0, fmt.Errorf("agentManagerRepository.ListAgentsPaged: %w", err)
	}
	return rows, total, nil
}

func (r *agentManagerRepository) ListAgentCommands(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListAgentsCommandsResponse, error) {
	resp, err := r.client.ListAgentCommands(ctx, searchQuery, pageNumber, pageSize, sortBy)
	if err != nil {
		return nil, fmt.Errorf("agentManagerRepository.ListAgentCommands: %w", err)
	}
	return resp, nil
}

func (r *agentManagerRepository) UpdateAgent(ctx context.Context, agentID uint32, agentKey, ip, hostname string) (*agent.AuthResponse, error) {
	resp, err := r.client.UpdateAgent(ctx, agentID, agentKey, ip, hostname)
	if err != nil {
		return nil, fmt.Errorf("agentManagerRepository.UpdateAgent: %w", err)
	}
	return resp, nil
}
