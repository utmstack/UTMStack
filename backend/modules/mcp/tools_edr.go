package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// ------------------------------------------------------------------
// edr.* — fleet EDR console: monitor endpoints and dispatch actions.
// Reads fold stored status reports (never wake the endpoint); writes are
// fire-and-forget — the endpoint's answer lands in the agent command history.
// ------------------------------------------------------------------

func registerEDR(m *Module) {
	uc := m.deps.EDR.Usecase()

	Add(m, &mcp.Tool{
		Name: "edr.overview", Title: "EDR fleet overview",
		Description: "Fleet summary: endpoints with the EDR module, healthy vs unhealthy, stale signatures, drifted endpoints, and detections in the last 24h.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "edr.read"},
		func(ctx context.Context, _ *authz.Actor, _ edrEmptyInput) (any, error) {
			return uc.Overview(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "edr.endpoints", Title: "List EDR endpoints",
		Description: "Per-endpoint EDR state: module presence, engine health, signature version, drift, and last report time.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "edr.read"},
		func(ctx context.Context, _ *authz.Actor, _ edrEmptyInput) (any, error) {
			return uc.Endpoints(ctx)
		})

	Add(m, &mcp.Tool{
		Name: "edr.endpoint", Title: "One EDR endpoint",
		Description: "Full stored EDR state for one endpoint.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "edr.read"},
		func(ctx context.Context, _ *authz.Actor, in edrEndpointInput) (any, error) {
			return uc.Endpoint(ctx, in.AgentID)
		})

	Add(m, &mcp.Tool{
		Name: "edr.command", Title: "Check a dispatched EDR action result",
		Description: "Poll the answer of a dispatched action by cmd_id: not found = the endpoint is offline or the dispatch failed; pending = still working; error with a refusal reason (unsupported_action, validation_error) = the endpoint answered and refused.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "edr.read"},
		func(ctx context.Context, _ *authz.Actor, in edrCommandInput) (any, error) {
			return uc.CommandResult(ctx, in.AgentID, in.CmdID)
		})

	Add(m, &mcp.Tool{
		Name: "edr.quarantine_list", Title: "Quarantine items across the fleet",
		Description: "Live quarantined files on all endpoints (or one, when agent_id is set), newest first.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "edr.read"},
		func(ctx context.Context, _ *authz.Actor, in edrEndpointFilterInput) (any, error) {
			var filter *uint32
			if in.AgentID > 0 {
				filter = &in.AgentID
			}
			return uc.Quarantine(ctx, filter)
		})

	Add(m, &mcp.Tool{
		Name: "edr.quarantine_action", Title: "Restore or purge a quarantined file",
		Description: "Act on one quarantined file on one endpoint: restore puts it back, purge permanently deletes the stored copy.",
	}, Gate{Permission: "edr.write"},
		func(ctx context.Context, a *authz.Actor, in edrQuarantineActionInput) (any, error) {
			if in.Op != "restore" && in.Op != "purge" {
				return nil, fmt.Errorf("op must be restore or purge, got %q", in.Op)
			}
			return uc.QuarantineItem(ctx, in.AgentID, in.ItemID, in.Op, a.Email)
		})

	Add(m, &mcp.Tool{
		Name: "edr.action", Title: "Dispatch an EDR action to an endpoint",
		Description: "Dispatch one action to a connected endpoint (scan_path, full_scan, kill_process, isolate, release_isolation, blocklist_refresh, enable, disable, status, quarantine_list). Fire-and-forget: the endpoint's answer is in its command history. Offline endpoints fail fast.",
	}, Gate{Permission: "edr.write"},
		func(ctx context.Context, a *authz.Actor, in edrActionInput) (any, error) {
			return uc.Action(ctx, in.AgentID, in.Action, string(in.Payload), a.Email, in.Reason)
		})
}

type edrEmptyInput struct{}

type edrEndpointInput struct {
	AgentID uint32 `json:"agent_id" jsonschema:"Endpoint agent id (from edr.endpoints)"`
}

type edrCommandInput struct {
	AgentID uint32 `json:"agent_id" jsonschema:"Endpoint agent id"`
	CmdID   string `json:"cmd_id" jsonschema:"Command id returned by edr.action or edr.quarantine_action"`
}

type edrEndpointFilterInput struct {
	AgentID uint32 `json:"agent_id,omitempty" jsonschema:"Filter to one endpoint; 0 = whole fleet"`
}

type edrQuarantineActionInput struct {
	AgentID uint32 `json:"agent_id" jsonschema:"Endpoint holding the quarantined file"`
	ItemID  string `json:"item_id" jsonschema:"Quarantine id (from edr.quarantine_list)"`
	Op      string `json:"op" jsonschema:"restore or purge"`
}

type edrActionInput struct {
	AgentID uint32          `json:"agent_id" jsonschema:"Target endpoint agent id"`
	Action  string          `json:"action" jsonschema:"Action verb (scan_path, full_scan, kill_process, isolate, release_isolation, blocklist_refresh, enable, disable, status, quarantine_list)"`
	Payload json.RawMessage `json:"payload,omitempty" jsonschema:"Action arguments as a JSON object (e.g. {\"path\": \"C:\\temp\"} for scan_path)"`
	Reason  string          `json:"reason,omitempty" jsonschema:"Why this action is being taken"`
}
