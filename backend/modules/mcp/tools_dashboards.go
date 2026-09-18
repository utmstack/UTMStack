package mcp

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerDashboards(m *Module) {
	registerDashboardDashboards(m)
	registerDashboardVisualizations(m)
}

// ---- dashboards.* ----------------------------------------------------------

type dashboardUpsertInput struct {
	ID          uuid.UUID `json:"id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Config      string    `json:"config,omitempty"`
}

type dashboardListInput struct {
	Name string `json:"name,omitempty"`
	Page int    `json:"page,omitempty"`
	Size int    `json:"size,omitempty"`
}

type dashboardIDInput struct {
	ID uuid.UUID `json:"id"`
}

func registerDashboardDashboards(m *Module) {
	uc := m.deps.Dashboards.GetDashboardUsecase()

	Add(m, &mcp.Tool{
		Name: "dashboards.create", Title: "Create dashboard",
		Description: "Creates an empty dashboard shell. Add its widgets afterwards with " +
			"visualizations.create, passing the returned id as dashboard_id.",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in dashboardUpsertInput) (any, error) {
			d := &domain.Dashboard{Name: in.Name, Description: in.Description, Config: in.Config}
			return uc.Create(ctx, d, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "dashboards.update", Title: "Update dashboard",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in dashboardUpsertInput) (any, error) {
			d := &domain.Dashboard{ID: in.ID, Name: in.Name, Description: in.Description, Config: in.Config}
			return uc.Update(ctx, d, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "dashboards.list", Title: "List dashboards",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.DashboardFilter{
				Name: in.Name, Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "dashboards.get", Title: "Get dashboard",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "dashboards.delete", Title: "Delete dashboard",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}

// ---- visualizations.* ------------------------------------------------------

type visualizationUpsertInput struct {
	ID          uuid.UUID `json:"id,omitempty"`
	DashboardID uuid.UUID `json:"dashboard_id"`
	Spec        string    `json:"spec"`
	Config      string    `json:"config,omitempty"`
	Layout      string    `json:"layout,omitempty"`
}

type visualizationListInput struct {
	DashboardID *uuid.UUID `json:"dashboard_id,omitempty"`
	Page        int        `json:"page,omitempty"`
	Size        int        `json:"size,omitempty"`
}

const visualizationSpecDoc = "Creates one chart widget on a dashboard, given its dashboard_id. " +
	"All three of spec/config/layout are JSON given as strings, not nested objects.\n\n" +
	"spec (required) is the question the widget asks. Fields: " +
	"dataset (required) is \"logs\" or \"alerts\" — nothing else is valid. " +
	"chart (required) is \"metric\" (one number), \"category\" (top values of a field), \"time\" (a series over time), or \"table\" (raw rows). " +
	"metric.agg must be the string \"count\" — the event store only counts records, no sum/avg/cardinality. " +
	"dimension is REQUIRED when chart is \"category\" (the field broken down by) and unused otherwise — " +
	"call store.dataset.fields on the dataset first to get the real field name; a guessed or misspelled one returns zero buckets, not an error. " +
	"interval (time charts only) is one of \"1m\",\"5m\",\"15m\",\"1h\",\"1d\",\"1w\". " +
	"columns (table charts only) lists the field paths to project. " +
	"filters is an optional array of {field, op, value}; op is one of " +
	"eq, not_eq, in, not_in, gt, gte, lt, lte, between, not_between, contains, not_contains, starts_with, not_starts_with, ends_with, not_ends_with, exists, not_exists " +
	"(exists/not_exists take no value). limit caps how many buckets/series/rows come back.\n" +
	"Example — alerts by severity: {\"dataset\":\"alerts\",\"chart\":\"category\",\"dimension\":\"severity\",\"metric\":{\"agg\":\"count\"}}\n\n" +
	"config is the ECharts option merged with the query's data. Pass \"{}\" for a sensible auto-built chart " +
	"(bar/line for category/time, a big number for metric, a grid for table) — never omit it or pass an empty string, " +
	"that renders as a visible error instead of a chart. A partial option (e.g. {\"color\":[...]} or a custom \"title\") is merged over the default.\n\n" +
	"layout is this widget's grid position: {\"x\":int,\"y\":int,\"w\":int,\"h\":int} on the dashboard's 12-column grid " +
	"(row height 50px). The default size is w:4 h:8 — three widgets fit per row at that width. " +
	"Give each widget on the same dashboard a different x/y so they don't overlap; increasing y for each new row is enough."

func registerDashboardVisualizations(m *Module) {
	uc := m.deps.Dashboards.GetVisualizationUsecase()

	Add(m, &mcp.Tool{
		Name: "visualizations.create", Title: "Create visualization",
		Description: visualizationSpecDoc,
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in visualizationUpsertInput) (any, error) {
			return uc.Create(ctx, &domain.Visualization{
				DashboardID: in.DashboardID,
				Spec:        in.Spec, Config: in.Config, Layout: in.Layout,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.update", Title: "Update visualization",
		Description: visualizationSpecDoc,
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in visualizationUpsertInput) (any, error) {
			return uc.Update(ctx, &domain.Visualization{
				ID: in.ID, DashboardID: in.DashboardID,
				Spec: in.Spec, Config: in.Config, Layout: in.Layout,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.list", Title: "List visualizations",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in visualizationListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.VisualizationFilter{
				DashboardID: in.DashboardID,
				Params:      database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.get", Title: "Get visualization",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.delete", Title: "Delete visualization",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
