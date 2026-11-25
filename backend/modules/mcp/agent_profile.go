package mcp

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

// group is a capability area of the agent. What the agent can see is its core plus the groups the
// caller turned on for this request, and nothing else exists for it.
type group string

const (
	// groupCore is always visible: what investigating an alert takes, and writing the assessment note.
	groupCore group = ""

	groupAlertActions    group = "alert_actions"    // change the status of alerts, tag them
	groupIncidentActions group = "incident_actions" // open incidents from alerts, move them, note them
	groupDashboards      group = "dashboards"
	groupCompliance      group = "compliance"
	groupSOAR            group = "soar"
	groupDetection       group = "detection"
	groupDatasources     group = "datasources"
)

// agentTools is every tool the SOC agent can ever see, with the group it belongs to. The agent is the
// internal caller of this endpoint (the soc-ai service, which presents the internal key), and the
// internal key passes every gate, so what it may do is decided here and not by permissions.
//
// It is short on purpose. A model handed a hundred small tools makes a hundred small calls and never
// gets to an answer; and one that can reach tenants, billing or users can do harm no alert justifies.
// So only what investigating and handling alerts takes is in the core (nine tools), and everything
// else is a group a caller has to ask for. Left out altogether: user, tenant, license, branding and
// settings management, bulk operations, raw SQL and CSV export, and anything that deletes. Duplicates
// are left out too: store.search already returns the total, and alerts.convert_to_incident already
// creates an incident or adds to one.
//
// Humans and other clients keep the full server; they are not affected by this list.
var agentTools = map[string]group{
	// Core: judge an alert, look around it, write the assessment.
	"alerts.score":          groupCore,
	"alerts.update_notes":   groupCore,
	"store.search":          groupCore,
	"store.property_values": groupCore,
	"store.dataset.fields":  groupCore,
	"incidents.get":         groupCore,
	"incidents.list":        groupCore,
	"incident_alerts.list":  groupCore,
	"incident_notes.list":   groupCore,

	"alerts.update_status": groupAlertActions,
	"alerts.update_tags":   groupAlertActions,

	"alerts.convert_to_incident": groupIncidentActions,
	"incidents.change_status":    groupIncidentActions,
	"incident_notes.create":      groupIncidentActions,

	"dashboards.list":       groupDashboards,
	"dashboards.get":        groupDashboards,
	"dashboards.create":     groupDashboards,
	"dashboards.update":     groupDashboards,
	"visualizations.list":   groupDashboards,
	"visualizations.query":  groupDashboards,
	"visualizations.create": groupDashboards,
	"visualizations.update": groupDashboards,

	"compliance.framework.list":  groupCompliance,
	"compliance.framework.get":   groupCompliance,
	"compliance.control.list":    groupCompliance,
	"compliance.report.evaluate": groupCompliance,

	"soar.rule.list":        groupSOAR,
	"soar.rule.get":         groupSOAR,
	"soar.rule.create":      groupSOAR,
	"soar.rule.update":      groupSOAR,
	"soar.rule.set_enabled": groupSOAR,
	"soar.execution.list":   groupSOAR,
	"soar.node_types":       groupSOAR,

	"correlation_rule.list":       groupDetection,
	"correlation_rule.get":        groupDetection,
	"correlation_rule.create":     groupDetection,
	"correlation_rule.update":     groupDetection,
	"correlation_rule.set_active": groupDetection,

	"datasources.list":       groupDatasources,
	"datasources.get":        groupDatasources,
	"ingestion_stats.totals": groupDatasources,
}

// knownGroups are the groups a caller may name.
var knownGroups = map[group]bool{
	groupAlertActions: true, groupIncidentActions: true, groupDashboards: true, groupCompliance: true,
	groupSOAR: true, groupDetection: true, groupDatasources: true,
}

// agentHeader carries the groups a request turns on, comma separated. It can only add groups from the
// list above: the agent never sees a tool outside agentTools, whatever the header says, and without it
// only the core exists.
const agentHeader = "X-Agent-Groups"

type agentGroupsKey struct{}

// WithAgentGroups returns ctx with the groups named in csv turned on. Unknown names are ignored.
func WithAgentGroups(ctx context.Context, csv string) context.Context {
	on := map[group]bool{}

	for _, part := range strings.Split(csv, ",") {
		if g := group(strings.TrimSpace(part)); knownGroups[g] {
			on[g] = true
		}
	}

	return context.WithValue(ctx, agentGroupsKey{}, on)
}

func agentGroupsFrom(ctx context.Context) map[group]bool {
	on, _ := ctx.Value(agentGroupsKey{}).(map[group]bool)

	return on
}

// visibleToAgent says whether a tool exists for the agent in a request that turned on these groups.
func visibleToAgent(name string, on map[group]bool) bool {
	g, ok := agentTools[name]

	return ok && (g == groupCore || on[g])
}

// agentFilter hides the tools a request did not turn on, both from the list and from a direct call, so
// a model cannot call what it was never shown.
func agentFilter(next mcp.MethodHandler) mcp.MethodHandler {
	return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
		on := agentGroupsFrom(ctx)

		switch method {
		case "tools/list":
			res, err := next(ctx, method, req)
			if err != nil {
				return res, err
			}

			if list, ok := res.(*mcp.ListToolsResult); ok {
				kept := make([]*mcp.Tool, 0, len(list.Tools))

				for _, t := range list.Tools {
					if visibleToAgent(t.Name, on) {
						kept = append(kept, t)
					}
				}

				list.Tools = kept
			}

			return res, nil

		case "tools/call":
			if p, ok := req.GetParams().(*mcp.CallToolParamsRaw); ok && !visibleToAgent(p.Name, on) {
				return nil, fmt.Errorf("the tool %q is not available", p.Name)
			}
		}

		return next(ctx, method, req)
	}
}

// liftAgentGroups reads the header for the internal caller. Anyone else cannot use it: they get the
// full server, which does not look at it.
func liftAgentGroups() gin.HandlerFunc {
	return func(c *gin.Context) {
		if actor := middleware.ActorFromGin(c); actor != nil && actor.Internal {
			c.Request = c.Request.WithContext(WithAgentGroups(c.Request.Context(), c.GetHeader(agentHeader)))
		}

		c.Next()
	}
}

// serverFor picks the server a request talks to: the agent's for the internal caller, the full one for
// everybody else. The choice comes from who the caller is, not from anything the caller sends, so a
// client cannot ask for more by leaving a header out.
func (m *Module) serverFor(r *http.Request) *mcp.Server {
	if m.agent != nil {
		if actor := ActorFromContext(r.Context()); actor != nil && actor.Internal {
			return m.agent
		}
	}

	return m.server
}
