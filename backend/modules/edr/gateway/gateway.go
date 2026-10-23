package gateway

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/edr/connectors"
	"github.com/utmstack/utmstack/backend/modules/edr/dto"
	am "github.com/utmstack/utmstack/backend/pkg/agentmanager"
)

// managerGateway adapts the agent-manager gRPC client to the EDR gateway.
type managerGateway struct {
	client *am.AgentManagerClient
}

func NewManagerGateway(client *am.AgentManagerClient) connectors.AgentGateway {
	return &managerGateway{client: client}
}

func (g *managerGateway) ListEndpoints(ctx context.Context) ([]dto.Endpoint, error) {
	rows, _, err := g.client.ListAgents(ctx, "")
	if err != nil {
		return nil, err
	}
	out := make([]dto.Endpoint, 0, len(rows))
	for _, a := range rows {
		out = append(out, dto.Endpoint{
			AgentID:  a.GetId(),
			Hostname: a.GetHostname(),
			IP:       a.GetIp(),
			OS:       a.GetOs(),
			Platform: a.GetPlatform(),
			Version:  a.GetVersion(),
			Status:   a.GetStatus().String(),
		})
	}
	return out, nil
}

func (g *managerGateway) EdrStatuses(ctx context.Context) ([]dto.EdrStatus, error) {
	rows, err := g.client.GetEdrStatuses(ctx, 0)
	if err != nil {
		return nil, err
	}
	out := make([]dto.EdrStatus, 0, len(rows))
	for _, r := range rows {
		var at time.Time
		if ts := r.GetReportedAt(); ts != nil {
			at = ts.AsTime()
		}
		out = append(out, dto.EdrStatus{
			AgentID:       r.GetAgentId(),
			StatusJSON:    r.GetStatusJson(),
			PolicyVersion: r.GetPolicyVersion(),
			ReportedAt:    at,
		})
	}
	return out, nil
}

func (g *managerGateway) Dispatch(ctx context.Context, agentID uint32, action, payloadJSON, executedBy, reason string) (string, error) {
	resp, err := g.client.DispatchEdrCommand(ctx, strconv.FormatUint(uint64(agentID), 10), "", action, payloadJSON, executedBy, reason)
	if err != nil {
		catcher.Error("edr dispatch failed", err, map[string]any{"agent_id": agentID, "action": action, "process": "edr"})
		return "", err
	}
	return resp.GetCmdId(), nil
}

func (g *managerGateway) Command(ctx context.Context, agentID uint32, cmdID string) (*dto.CommandRow, error) {
	row, err := g.client.GetEdrCommand(ctx, agentID, cmdID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return &dto.CommandRow{
		CmdID:      row.GetCmdId(),
		Action:     strings.TrimPrefix(row.GetCommand(), "edr: "),
		Status:     row.GetCommandStatus().String(),
		Result:     row.GetResult(),
		ExecutedBy: row.GetExecutedBy(),
	}, nil
}
