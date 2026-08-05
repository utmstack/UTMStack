package agentmanager

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

type spyAgentService struct {
	agent.AgentServiceClient
	saw *agent.ListRequest
}

func (s *spyAgentService) ListAgents(_ context.Context, in *agent.ListRequest, _ ...grpc.CallOption) (*agent.ListAgentsResponse, error) {
	s.saw = in
	return &agent.ListAgentsResponse{}, nil
}

func (s *spyAgentService) ListAgentCommands(_ context.Context, in *agent.ListRequest, _ ...grpc.CallOption) (*agent.ListAgentsCommandsResponse, error) {
	s.saw = in
	return &agent.ListAgentsCommandsResponse{}, nil
}

type spyPanelService struct {
	agent.PanelServiceClient
	saw *agent.ConnectionKeyRequest
}

func (s *spyPanelService) GetConnectionKey(_ context.Context, in *agent.ConnectionKeyRequest, _ ...grpc.CallOption) (*agent.ConnectionKeyResponse, error) {
	s.saw = in
	return &agent.ConnectionKeyResponse{}, nil
}

func (s *spyPanelService) RotateConnectionKey(_ context.Context, in *agent.ConnectionKeyRequest, _ ...grpc.CallOption) (*agent.ConnectionKeyResponse, error) {
	s.saw = in
	return &agent.ConnectionKeyResponse{}, nil
}

type spyCollectorService struct {
	agent.CollectorServiceClient
	saw *agent.ListRequest
}

func (s *spyCollectorService) ListCollector(_ context.Context, in *agent.ListRequest, _ ...grpc.CallOption) (*agent.ListCollectorResponse, error) {
	s.saw = in
	return &agent.ListCollectorResponse{}, nil
}

// A hostname is unique inside a tenant, not across them. Scoring reads the
// asset details of whatever comes back and SOAR sends it a response action, so
// an unscoped lookup acts on another tenant's machine whenever two of them run
// a host of the same name — which is most of the time.
func TestListingsCarryTheActingTenant(t *testing.T) {
	agents := &spyAgentService{}
	collectors := &spyCollectorService{}
	c := &AgentManagerClient{agentService: agents, collectorService: collectors}

	ctx := authz.WithTenantID(context.Background(), customerTenant)

	if _, _, err := c.ListAgents(ctx, "hostname.Is=DC01"); err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	if agents.saw.GetTenantId() != customerTenant {
		t.Errorf("ListAgents tenant = %q, want %q", agents.saw.GetTenantId(), customerTenant)
	}

	if _, _, err := c.ListAgentsPaged(ctx, "", 1, 10, ""); err != nil {
		t.Fatalf("ListAgentsPaged: %v", err)
	}
	if agents.saw.GetTenantId() != customerTenant {
		t.Errorf("ListAgentsPaged tenant = %q, want %q", agents.saw.GetTenantId(), customerTenant)
	}

	if _, err := c.ListAgentCommands(ctx, "", 1, 10, ""); err != nil {
		t.Fatalf("ListAgentCommands: %v", err)
	}
	if agents.saw.GetTenantId() != customerTenant {
		t.Errorf("ListAgentCommands tenant = %q, want %q", agents.saw.GetTenantId(), customerTenant)
	}

	if _, err := c.ListCollectors(ctx, "", 1, 10, ""); err != nil {
		t.Fatalf("ListCollectors: %v", err)
	}
	if collectors.saw.GetTenantId() != customerTenant {
		t.Errorf("ListCollectors tenant = %q, want %q", collectors.saw.GetTenantId(), customerTenant)
	}
}

// No tenant asks for every one, which is what an install that is not
// multi-tenant wants and what these calls did before.
func TestNoTenantAsksForEveryOne(t *testing.T) {
	agents := &spyAgentService{}
	c := &AgentManagerClient{agentService: agents}

	if _, _, err := c.ListAgents(context.Background(), ""); err != nil {
		t.Fatalf("ListAgents: %v", err)
	}
	if agents.saw.GetTenantId() != "" {
		t.Errorf("tenant = %q, want it empty", agents.saw.GetTenantId())
	}
}

// The connection key is what enrols an agent, so it decides which tenant that
// agent reports into. An empty request makes the agent-manager answer with the
// default tenant's key: every tenant would enrol into the operator's, and
// rotating would revoke enrolment for everyone at once.
func TestConnectionKeyIsPerTenant(t *testing.T) {
	panel := &spyPanelService{}
	c := &AgentManagerClient{panelService: panel}

	ctx := authz.WithTenantID(context.Background(), customerTenant)

	if _, err := c.GetConnectionKey(ctx); err != nil {
		t.Fatalf("GetConnectionKey: %v", err)
	}
	if panel.saw.GetTenantId() != customerTenant {
		t.Errorf("Get tenant = %q, want %q", panel.saw.GetTenantId(), customerTenant)
	}

	if _, err := c.RotateConnectionKey(ctx); err != nil {
		t.Fatalf("RotateConnectionKey: %v", err)
	}
	if panel.saw.GetTenantId() != customerTenant {
		t.Errorf("Rotate tenant = %q, want %q", panel.saw.GetTenantId(), customerTenant)
	}
}
