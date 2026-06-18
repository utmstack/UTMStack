package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/adaudit"
	"github.com/utmstack/utmstack/backend/modules/alerts"
	"github.com/utmstack/utmstack/backend/modules/alertscoring"
	"github.com/utmstack/utmstack/backend/modules/appconfig"
	"github.com/utmstack/utmstack/backend/modules/audit"
	"github.com/utmstack/utmstack/backend/modules/billing"
	"github.com/utmstack/utmstack/backend/modules/compliance"
	"github.com/utmstack/utmstack/backend/modules/dashboards"
	"github.com/utmstack/utmstack/backend/modules/datasources"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing"
	"github.com/utmstack/utmstack/backend/modules/iam"
	"github.com/utmstack/utmstack/backend/modules/incidents"
	"github.com/utmstack/utmstack/backend/modules/integrations"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer"
	"github.com/utmstack/utmstack/backend/modules/notifications"
	opensearchgw "github.com/utmstack/utmstack/backend/modules/opensearch"
	"github.com/utmstack/utmstack/backend/modules/soar"
	"github.com/utmstack/utmstack/backend/modules/socai"
)

// Deps holds the other modules the MCP layer borrows usecases from. New
// dependencies are added here so module.go stays the single touch-point.
type Deps struct {
	IAM             *iam.Module
	Alerts          *alerts.Module
	AlertScoring    *alertscoring.Module
	Incidents       *incidents.Module
	SOAR            *soar.Module
	Compliance      *compliance.Module
	Audit           *audit.Module
	Dashboards      *dashboards.Module
	LogAnalyzer     *loganalyzer.Module
	OpenSearch      *opensearchgw.Module
	EventProcessing *eventprocessing.Module
	Datasources     *datasources.Module
	Integrations    *integrations.Module
	Notifications   *notifications.Module
	ADAudit         *adaudit.Module
	SOCAI           *socai.Module
	Billing         *billing.Module
	AppConfig       *appconfig.Module

	// ServerName / ServerVersion populate the MCP Implementation block sent
	// during the initialize handshake; clients display these in their UI.
	ServerName    string
	ServerVersion string
}

// Module is the MCP module. Build it with NewModule, then mount via
// RegisterRoutes. The wrapped *mcp.Server is built once at startup; tool
// registrations are idempotent within a single buildServer call.
type Module struct {
	deps      *Deps
	server    *mcp.Server
	toolCount int
}

// NewModule constructs the module and registers every tool, prompt, and
// resource against a freshly built SDK server. It panics on schema-inference
// errors (which would indicate a programmer bug, not a runtime fault).
func NewModule(deps *Deps) *Module {
	if deps == nil {
		panic("mcp.NewModule: Deps must not be nil")
	}
	m := &Module{deps: deps}
	m.server = m.buildServer()
	return m
}

// Server returns the wrapped *mcp.Server. Exposed mainly for tests that want
// to drive the server in-process without going through HTTP.
func (m *Module) Server() *mcp.Server { return m.server }

// ToolCount returns the number of tools registered. The SDK doesn't expose a
// public reader, so we maintain a running tally inside addTool.
func (m *Module) ToolCount() int { return m.toolCount }
