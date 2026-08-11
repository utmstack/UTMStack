package mcp

import (
	"context"
	"errors"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// buildServer constructs the *mcp.Server and registers every tool, prompt,
// and resource. Called exactly once from NewModule.
func (m *Module) buildServer() *mcp.Server {
	name := m.deps.ServerName
	if name == "" {
		name = "utmstack"
	}
	version := m.deps.ServerVersion
	if version == "" {
		version = "v1"
	}
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    name,
		Title:   "UTMStack",
		Version: version,
	}, nil)
	m.server = srv

	registerAlerts(m)
	registerAlertScoring(m)
	registerIncidents(m)
	// registerSOAR(m) // temporarily disabled — SOAR response actions not exposed to the agent for now
	registerCompliance(m)
	registerLogAnalyzer(m)
	registerDashboards(m)
	registerEventProcessing(m)
	registerDatasources(m)
	// registerIntegrations(m) // temporarily disabled — integration management not exposed for now
	// registerIAM(m) // temporarily disabled — user/role management not exposed for now
	registerAudit(m)
	registerADAudit(m)
	// registerNotifications(m) // temporarily disabled — notification management not exposed for now
	registerAppConfig(m)
	registerBilling(m)
	registerSOCAI(m)
	registerTenants(m)

	registerResources(m)

	catcher.Info(fmt.Sprintf("MCP server built: %d tools registered", m.toolCount), nil)
	return srv
}

// Gate declares the authorization checks a tool requires. The zero Gate means
// "authenticated only" — every MCP request must already carry a valid actor
// because RegisterRoutes mounts the handler behind userAuth.
type Gate struct {
	// Permission, when non-empty, requires actor.Permissions to contain it
	// (mirrors middleware.RequirePermission).
	Permission string
	// Role, when non-empty, requires actor.Roles to contain it.
	Role string
	// Admin enforces the ROLE_ADMIN gate (same as middleware.RequireAdmin).
	Admin bool
	// Platform restricts the tool to the super-admin — an admin of the
	// default tenant (mirrors middleware.RequirePlatform). Use it for
	// tools that own the fleet: tenant CRUD, cross-tenant maintenance.
	Platform bool
	// Enterprise requires the active license to be Enterprise.
	Enterprise bool
	// MSSP requires the active license to be MSSP.
	MSSP bool
}

func (g Gate) check(actor *authz.Actor, flags licenseFlags) error {
	if actor == nil {
		return authz.ErrUnauthenticated
	}
	if actor.Internal {
		return nil
	}
	if g.Permission != "" {
		if err := authz.RequirePermission(actor, g.Permission); err != nil {
			return err
		}
	}
	if g.Role != "" {
		if err := authz.RequireRole(actor, g.Role); err != nil {
			return err
		}
	}
	if g.Admin {
		if err := authz.RequireAdmin(actor); err != nil {
			return err
		}
	}
	if g.Platform {
		if err := authz.RequirePlatform(actor); err != nil {
			return err
		}
	}
	if g.Enterprise {
		if err := authz.RequireEnterprise(actor, flags.isEnterprise); err != nil {
			return err
		}
	}
	if g.MSSP {
		if err := authz.RequireMSSP(actor, flags.isMSSP); err != nil {
			return err
		}
	}
	return nil
}

// Handler is the inner handler tool files implement. It receives the validated
// actor and the typed input; the wrapper takes care of auth, error mapping,
// and audit logging.
type Handler[In, Out any] func(ctx context.Context, actor *authz.Actor, in In) (Out, error)

// Add registers one tool against the MCP server with the standard
// gate-check + audit-log envelope. Each tools_*.go file calls this once per
// usecase method it exposes.
//
// Out should typically be `any` so the SDK serializes whatever the usecase
// returns; for tools that have no useful output (mutations) return nil.
func Add[In, Out any](m *Module, t *mcp.Tool, gate Gate, h Handler[In, Out]) {
	m.toolCount++
	toolName := t.Name
	mcp.AddTool(m.server, t, func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Out, error) {
		var zero Out
		actor := ActorFromContext(ctx)
		if actor == nil {
			return errorResult("authentication required: MCP tools must be invoked under a JWT or Utm-Api-Key"), zero, authz.ErrUnauthenticated
		}
		if err := gate.check(actor, m.deps.licenseFlags()); err != nil {
			m.auditDenied(ctx, actor, toolName, err)
			return errorResult("permission denied: " + err.Error()), zero, err
		}
		out, err := h(ctx, actor, in)
		if err != nil {
			m.auditFailure(ctx, actor, toolName, err)
			return errorResult(err.Error()), zero, err
		}
		m.auditSuccess(ctx, actor, toolName)
		return nil, out, nil
	})
}

// errorResult builds a CallToolResult with IsError set and a TextContent
// message. The SDK still returns the Go error separately, but spec clients
// surface the IsError flag and text to the user.
func errorResult(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}
}

// IsAuthErr is true for any authz error class — handy for tests.
func IsAuthErr(err error) bool {
	return errors.Is(err, authz.ErrUnauthenticated) ||
		errors.Is(err, authz.ErrMissingPermission) ||
		errors.Is(err, authz.ErrMissingRole) ||
		errors.Is(err, authz.ErrEnterpriseRequired) ||
		errors.Is(err, authz.ErrMSSPRequired) ||
		errors.Is(err, authz.ErrInternalOnly)
}
