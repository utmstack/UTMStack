package repository

import (
	"context"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	dsconnectors "github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager"
	"github.com/utmstack/utmstack/backend/pkg/agentmanager/agent"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type assetLookup struct {
	agents *agentmanager.AgentManagerClient // may be nil when agent-manager is unreachable
	ds     dsconnectors.DatasourceUsecase   // may be nil
}

func NewAssetLookup(agents *agentmanager.AgentManagerClient, ds dsconnectors.DatasourceUsecase) connectors.AssetLookup {
	return &assetLookup{agents: agents, ds: ds}
}

func (a *assetLookup) Lookup(ctx context.Context, hostname string) connectors.AssetInfo {
	info := connectors.AssetInfo{Hostname: hostname}
	if strings.TrimSpace(hostname) == "" {
		return info
	}
	a.enrichFromAgent(ctx, hostname, &info)
	a.enrichFromDatasource(ctx, hostname, &info)
	return info
}

func (a *assetLookup) enrichFromAgent(ctx context.Context, hostname string, info *connectors.AssetInfo) {
	if a.agents == nil {
		return
	}
	rows, _, err := a.agents.ListAgents(ctx, "hostname.Is="+hostname)
	if err != nil || len(rows) == 0 {
		return
	}
	ag := rows[0]
	info.OS = ag.GetOs()
	info.OSVersion = strings.Trim(ag.GetOsMajorVersion()+"."+ag.GetOsMinorVersion(), ".")
	info.Status = agentStatus(ag.GetStatus())

	for _, raw := range strings.Split(ag.GetAddresses(), ",") {
		p := strings.TrimSpace(raw)
		if p == "" {
			continue
		}
		info.IPs = append(info.IPs, strings.SplitN(p, "/", 2)[0])
	}
	if len(info.IPs) == 0 && ag.GetIp() != "" {
		info.IPs = append(info.IPs, ag.GetIp())
	}
}

func (a *assetLookup) enrichFromDatasource(ctx context.Context, hostname string, info *connectors.AssetInfo) {
	if a.ds == nil {
		return
	}
	res, err := a.ds.List(ctx, common_models.ListRequest{
		PageNumber:  1,
		PageSize:    5,
		SearchQuery: "name.contains." + hostname,
	})
	if err != nil {
		return
	}
	for _, d := range res.Items {
		if !strings.EqualFold(d.Name, hostname) {
			continue
		}
		info.Confidentiality = d.AssetConfidentiality
		info.Integrity = d.AssetIntegrity
		info.Availability = d.AssetAvailability
		info.HasSensitivity = true
		return
	}
}

func agentStatus(s agent.Status) string {
	switch s {
	case agent.Status_ONLINE:
		return "ONLINE"
	case agent.Status_OFFLINE:
		return "OFFLINE"
	default:
		return "UNKNOWN"
	}
}
