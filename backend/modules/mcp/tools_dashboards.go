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
	// Spec is the question the widget asks, as JSON: dataset, chart,
	// aggregation, breakdown, filters. It replaced the SQL a visualization used
	// to carry.
	Spec   string `json:"spec"`
	Config string `json:"config,omitempty"`
	Layout string `json:"layout,omitempty"`
}

type visualizationListInput struct {
	DashboardID *uuid.UUID `json:"dashboard_id,omitempty"`
	Page        int        `json:"page,omitempty"`
	Size        int        `json:"size,omitempty"`
}

func registerDashboardVisualizations(m *Module) {
	uc := m.deps.Dashboards.GetVisualizationUsecase()

	Add(m, &mcp.Tool{
		Name: "visualizations.create", Title: "Create visualization",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in visualizationUpsertInput) (any, error) {
			return uc.Create(ctx, &domain.Visualization{
				DashboardID: in.DashboardID,
				Spec:        in.Spec, Config: in.Config, Layout: in.Layout,
			}, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.update", Title: "Update visualization",
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
