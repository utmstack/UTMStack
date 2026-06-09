package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/modules/dashboards/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerDashboards(m *Module) {
	registerDashboardDashboards(m)
	registerDashboardVisualizations(m)
	registerDashboardLayouts(m)
}

// ---- dashboards.* ----------------------------------------------------------

type dashboardUpsertInput struct {
	ID          uint64 `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Config      string `json:"config,omitempty"`
}

type dashboardListInput struct {
	Name string `json:"name,omitempty"`
	Page int    `json:"page,omitempty"`
	Size int    `json:"size,omitempty"`
}

type dashboardIDInput struct {
	ID uint64 `json:"id"`
}

func registerDashboardDashboards(m *Module) {
	uc := m.deps.Dashboards.GetDashboardUsecase()

	Add(m, &mcp.Tool{
		Name: "dashboards.create", Title: "Create dashboard",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in dashboardUpsertInput) (any, error) {
			d := &domain.Dashboard{Name: in.Name, Description: in.Description, Config: in.Config}
			return uc.Create(ctx, d, actor.Login)
		})

	Add(m, &mcp.Tool{
		Name: "dashboards.update", Title: "Update dashboard",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in dashboardUpsertInput) (any, error) {
			d := &domain.Dashboard{ID: in.ID, Name: in.Name, Description: in.Description, Config: in.Config}
			return uc.Update(ctx, d, actor.Login)
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
	ID          uint64 `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	SQLQuery    string `json:"sql_query,omitempty"`
	Config      string `json:"config,omitempty"`
}

func registerDashboardVisualizations(m *Module) {
	uc := m.deps.Dashboards.GetVisualizationUsecase()

	Add(m, &mcp.Tool{
		Name: "visualizations.create", Title: "Create visualization",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in visualizationUpsertInput) (any, error) {
			return uc.Create(ctx, &domain.Visualization{
				Name: in.Name, Description: in.Description, SQLQuery: in.SQLQuery, Config: in.Config,
			}, actor.Login)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.update", Title: "Update visualization",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, actor *authz.Actor, in visualizationUpsertInput) (any, error) {
			return uc.Update(ctx, &domain.Visualization{
				ID: in.ID, Name: in.Name, Description: in.Description, SQLQuery: in.SQLQuery, Config: in.Config,
			}, actor.Login)
		})

	Add(m, &mcp.Tool{
		Name: "visualizations.list", Title: "List visualizations",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.VisualizationFilter{
				Name: in.Name, Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
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

// ---- dashboard_layouts.* ---------------------------------------------------

type dashboardLayoutUpsertInput struct {
	ID              uint64 `json:"id,omitempty"`
	IDDashboard     uint64 `json:"id_dashboard"`
	IDVisualization uint64 `json:"id_visualization"`
	Layout          string `json:"layout"`
}

type dashboardLayoutListInput struct {
	IDDashboard     *uint64 `json:"id_dashboard,omitempty"`
	IDVisualization *uint64 `json:"id_visualization,omitempty"`
	Page            int     `json:"page,omitempty"`
	Size            int     `json:"size,omitempty"`
}

func registerDashboardLayouts(m *Module) {
	uc := m.deps.Dashboards.GetLayoutUsecase()

	Add(m, &mcp.Tool{
		Name: "dashboard_layouts.create", Title: "Create dashboard layout",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, _ *authz.Actor, in dashboardLayoutUpsertInput) (any, error) {
			return uc.Create(ctx, &domain.DashboardVisualization{
				IDDashboard: in.IDDashboard, IDVisualization: in.IDVisualization, Layout: in.Layout,
			})
		})

	Add(m, &mcp.Tool{
		Name: "dashboard_layouts.update", Title: "Update dashboard layout",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, _ *authz.Actor, in dashboardLayoutUpsertInput) (any, error) {
			return uc.Update(ctx, &domain.DashboardVisualization{
				ID: in.ID, IDDashboard: in.IDDashboard, IDVisualization: in.IDVisualization, Layout: in.Layout,
			})
		})

	Add(m, &mcp.Tool{
		Name: "dashboard_layouts.list", Title: "List dashboard layouts",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardLayoutListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.LayoutFilter{
				IDDashboard: in.IDDashboard, IDVisualization: in.IDVisualization,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "dashboard_layouts.get", Title: "Get dashboard layout",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "dashboards.read"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "dashboard_layouts.delete", Title: "Delete dashboard layout",
	}, Gate{Permission: "dashboards.write"},
		func(ctx context.Context, _ *authz.Actor, in dashboardIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
