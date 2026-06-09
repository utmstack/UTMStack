package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
)

// registerResources is invoked from buildServer to mount every MCP resource.
// Resources are read-only documents the client can pre-load without spending
// tool-call budget — useful as context for the LLM.
func registerResources(m *Module) {
	registerCatalogResources(m)
	registerDocResources(m)
}

// ---- mcp://utmstack/catalog/* — live, server-backed catalogs ---------------

func registerCatalogResources(m *Module) {
	srv := m.server

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/catalog/integrations",
		Name:        "Integrations catalog",
		Description: "Live list of integration modules with data type, pretty name, category, and active flag. Useful before integrations.activate.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		items, total, err := m.deps.Integrations.Modules().List(ctx, connectors.ModuleListFilter{})
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, map[string]any{"items": items, "total": total})
	})

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/catalog/frameworks",
		Name:        "Compliance frameworks catalog",
		Description: "File-backed compliance framework catalog (key, name, enabled). Useful before compliance.report.generate.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		fs := m.deps.Compliance.GetFrameworkUsecase().ListFrameworks(ctx)
		return jsonResource(req.Params.URI, fs)
	})

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/catalog/alert-tags",
		Name:        "Alert tag catalog",
		Description: "User-defined alert tags. Useful for prompting the LLM with the existing taxonomy before alerts.update_tags.",
		MIMEType:    "application/json",
	}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		// No filter — pull up to the default page; clients that need more can use
		// the alert_tags.list tool with explicit pagination.
		items, total, err := m.deps.Alerts.GetAlertTagUsecase().List(ctx, alertTagListAll())
		if err != nil {
			return nil, err
		}
		return jsonResource(req.Params.URI, map[string]any{"items": items, "total": total})
	})
}

// ---- mcp://utmstack/docs/* — static reference docs -------------------------

func registerDocResources(m *Module) {
	srv := m.server

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/docs/filter-operators",
		Name:        "FilterType operator reference",
		Description: "The structured filter DSL operators used by opensearch.search, alerts.search_adversary, loganalyzer.*, dashboards, alert_tag_rules, and event-processing tools.",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return textResource(req.Params.URI, "text/markdown", filterOperatorsDoc)
	})

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/docs/event-types",
		Name:        "Audit event-type enum",
		Description: "List of ApplicationEventType values written to the audit log by REST handlers and MCP tools.",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return textResource(req.Params.URI, "text/markdown", eventTypesDoc)
	})

	srv.AddResource(&mcp.Resource{
		URI:         "mcp://utmstack/docs/catalog",
		Name:        "Full MCP tool catalog",
		Description: "Same content as the modules/mcp/catalog.json reference file — useful for AI clients that want to introspect every tool the server exposes.",
		MIMEType:    "application/json",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return textResource(req.Params.URI, "application/json", catalogJSON)
	})
}

// ---- helpers ---------------------------------------------------------------

func jsonResource(uri string, payload any) (*mcp.ReadResourceResult, error) {
	buf, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal resource %s: %w", uri, err)
	}
	return textResource(uri, "application/json", string(buf))
}

func textResource(uri, mime, text string) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{URI: uri, MIMEType: mime, Text: text},
		},
	}, nil
}
