package usecase

import (
	"context"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/network_scan/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

// agentGateway adapts the cross-module agentmanager.AgentManagerClient to the
// narrow connectors.AgentGateway interface that the network_scan usecase consumes.
type agentGateway struct {
	client *agentmanager.AgentManagerClient
}

func NewAgentGateway(client *agentmanager.AgentManagerClient) connectors.AgentGateway {
	return &agentGateway{client: client}
}

func (g *agentGateway) DeleteAgentByName(ctx context.Context, name string) error {
	if g.client == nil {
		return nil
	}
	agents, _, err := g.client.ListAgents(ctx, "hostname.Is="+name)
	if err != nil {
		return err
	}
	for _, a := range agents {
		if a.GetHostname() != name {
			continue
		}
		_, err := g.client.DeleteAgent(ctx, a.GetId(), a.GetAgentKey())
		if err != nil {
			return fmt.Errorf("agentGateway: DeleteAgent %q: %w", name, err)
		}
	}
	return nil
}

func (g *agentGateway) ListAgentNames(ctx context.Context) ([]string, error) {
	if g.client == nil {
		return nil, nil
	}
	agents, _, err := g.client.ListAgents(ctx, "")
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(agents))
	for _, a := range agents {
		if h := a.GetHostname(); h != "" {
			names = append(names, h)
		}
	}
	return names, nil
}
