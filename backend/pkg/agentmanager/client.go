package agentmanager

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"github.com/utmstack/utmstack/backend/pkg/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type AgentManagerClient struct {
	conn             *grpc.ClientConn
	agentService     agent.AgentServiceClient
	panelService     agent.PanelServiceClient
	collectorService agent.CollectorServiceClient
}

func NewClient() (*AgentManagerClient, error) {
	host := env.String("GRPC_AGENT_MANAGER_HOST", "localhost", false)
	port := env.String("GRPC_AGENT_MANAGER_PORT", "9000", false)
	internalKey := env.String("INTERNAL_KEY", "", false)

	addr := fmt.Sprintf("%s:%s", host, port)

	tlsCreds := credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true,
	})

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(tlsCreds),
		grpc.WithPerRPCCredentials(&internalKeyCreds{key: internalKey}),
	)
	if err != nil {
		return nil, fmt.Errorf("agentmanager: dial %s: %w", addr, err)
	}

	return &AgentManagerClient{
		conn:             conn,
		agentService:     agent.NewAgentServiceClient(conn),
		panelService:     agent.NewPanelServiceClient(conn),
		collectorService: agent.NewCollectorServiceClient(conn),
	}, nil
}

func (c *AgentManagerClient) Close() error {
	return c.conn.Close()
}

// ListAgents returns all matching agents together with the total count.
// searchQuery format: "hostname.Is=X&platform.Is=Y" (exact Java parity).
func (c *AgentManagerClient) ListAgents(ctx context.Context, searchQuery string) ([]*agent.Agent, int32, error) {
	resp, err := c.agentService.ListAgents(ctx, &agent.ListRequest{
		SearchQuery: searchQuery,
		PageNumber:  1,
		PageSize:    1000000, // matches Java AgentGrpcService.setPageSize(1000000)
	})
	if err != nil {
		return nil, 0, fmt.Errorf("agentmanager: ListAgents: %w", err)
	}
	return resp.GetRows(), resp.GetTotal(), nil
}

// ListAgentsPaged returns agents with explicit pagination parameters.
func (c *AgentManagerClient) ListAgentsPaged(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) ([]*agent.Agent, int32, error) {
	resp, err := c.agentService.ListAgents(ctx, &agent.ListRequest{
		SearchQuery: searchQuery,
		PageNumber:  pageNumber,
		PageSize:    pageSize,
		SortBy:      sortBy,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("agentmanager: ListAgentsPaged: %w", err)
	}
	return resp.GetRows(), resp.GetTotal(), nil
}

// ListCollectors calls gRPC ListCollector with searchQuery built from optional hostname/module params.
// searchQuery format: "module.Is=X&hostname.Is=Y" (exact Java parity).
func (c *AgentManagerClient) ListCollectors(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListCollectorResponse, error) {
	resp, err := c.collectorService.ListCollector(ctx, &agent.ListRequest{
		SearchQuery: searchQuery,
		PageNumber:  pageNumber,
		PageSize:    pageSize,
		SortBy:      sortBy,
	})
	if err != nil {
		return nil, fmt.Errorf("agentmanager: ListCollectors: %w", err)
	}
	return resp, nil
}

// ListAgentCommands calls gRPC ListAgentCommands with pagination.
func (c *AgentManagerClient) ListAgentCommands(ctx context.Context, searchQuery string, pageNumber, pageSize int32, sortBy string) (*agent.ListAgentsCommandsResponse, error) {
	resp, err := c.agentService.ListAgentCommands(ctx, &agent.ListRequest{
		SearchQuery: searchQuery,
		PageNumber:  pageNumber,
		PageSize:    pageSize,
		SortBy:      sortBy,
	})
	if err != nil {
		return nil, fmt.Errorf("agentmanager: ListAgentCommands: %w", err)
	}
	return resp, nil
}

// UpdateAgent calls gRPC UpdateAgent with per-RPC metadata headers key+id.
// Java: AgentGrpcService builds Metadata{key: agent.agentKey, id: agent.id} via MetadataUtils.
func (c *AgentManagerClient) UpdateAgent(ctx context.Context, agentID uint32, agentKey, ip, hostname string) (*agent.AuthResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx,
		"key", agentKey,
		"id", fmt.Sprintf("%d", agentID),
	)
	resp, err := c.agentService.UpdateAgent(ctx, &agent.AgentRequest{
		Ip:       ip,
		Hostname: hostname,
	})
	if err != nil {
		return nil, fmt.Errorf("agentmanager: UpdateAgent: %w", err)
	}
	return resp, nil
}

// DeleteAgent calls gRPC DeleteAgent with per-RPC metadata headers key+id+type=agent.
func (c *AgentManagerClient) DeleteAgent(ctx context.Context, agentID uint32, agentKey string) (*agent.AuthResponse, error) {
	ctx = metadata.AppendToOutgoingContext(ctx,
		"key", agentKey,
		"id", fmt.Sprintf("%d", agentID),
		"type", "agent",
	)
	resp, err := c.agentService.DeleteAgent(ctx, &agent.DeleteRequest{})
	if err != nil {
		return nil, fmt.Errorf("agentmanager: DeleteAgent: %w", err)
	}
	return resp, nil
}

// ProcessCommandStream opens a bidi gRPC stream to ProcessCommand, sends one
// UtmCommand (without calling CloseSend so the server keeps streaming), and
// returns a channel that emits each CommandResult chunk. The caller must
// drain both channels. The resultCh is closed after all chunks are received or
// an error occurs. errCh receives at most one error and is then closed.
// Cancelling ctx cancels the gRPC stream and causes the goroutine to exit.
func (c *AgentManagerClient) ProcessCommandStream(
	ctx context.Context,
	cmd *agent.UtmCommand,
) (<-chan *agent.CommandResult, <-chan error) {
	resultCh := make(chan *agent.CommandResult, 32)
	errCh := make(chan error, 1)

	go func() {
		defer close(resultCh)
		defer close(errCh)

		stream, err := c.panelService.ProcessCommand(ctx)
		if err != nil {
			errCh <- fmt.Errorf("agentmanager: ProcessCommandStream open: %w", err)
			return
		}
		if err := stream.Send(cmd); err != nil {
			errCh <- fmt.Errorf("agentmanager: ProcessCommandStream send: %w", err)
			return
		}
		// Do NOT call CloseSend — parity with Java (onCompleted commented out).
		for {
			result, recvErr := stream.Recv()
			if recvErr == io.EOF {
				return
			}
			if recvErr != nil {
				errCh <- fmt.Errorf("agentmanager: ProcessCommandStream recv: %w", recvErr)
				return
			}
			select {
			case resultCh <- result:
			case <-ctx.Done():
				return
			}
		}
	}()

	return resultCh, errCh
}

func (c *AgentManagerClient) ProcessCommand(ctx context.Context, cmd *agent.UtmCommand) (*agent.CommandResult, error) {
	stream, err := c.panelService.ProcessCommand(ctx)
	if err != nil {
		return nil, fmt.Errorf("agentmanager: ProcessCommand open stream: %w", err)
	}

	if err := stream.Send(cmd); err != nil {
		return nil, fmt.Errorf("agentmanager: ProcessCommand send: %w", err)
	}
	if err := stream.CloseSend(); err != nil {
		return nil, fmt.Errorf("agentmanager: ProcessCommand close send: %w", err)
	}

	result, err := stream.Recv()
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("agentmanager: ProcessCommand recv: %w", err)
	}
	return result, nil
}
