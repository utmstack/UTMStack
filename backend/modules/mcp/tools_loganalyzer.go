package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	ladomain "github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
)

func registerLogAnalyzer(m *Module) {
	registerLogAnalyzerAnalyzer(m)
	registerLogAnalyzerQueries(m)
}

// ---- loganalyzer.* (analyzer) ----------------------------------------------

type laTopValuesInput struct {
	Dataset  string                     `json:"dataset"`
	DataType string                     `json:"data_type,omitempty"`
	Field    string                     `json:"field"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
	Top      int                        `json:"top"`
}

func registerLogAnalyzerAnalyzer(m *Module) {
	uc := m.deps.LogAnalyzer.GetAnalyzerUsecase()

	Add(m, &mcp.Tool{
		Name: "loganalyzer.top_values", Title: "Top-N values for a field",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "loganalyzer.read"},
		func(ctx context.Context, _ *authz.Actor, in laTopValuesInput) (any, error) {
			top := in.Top
			if top <= 0 {
				top = 10
			}
			return uc.TopValues(ctx, in.Dataset, in.DataType, in.Field, in.Filters, top)
		})

	Add(m, &mcp.Tool{
		Name: "loganalyzer.chart_view", Title: "Chart view aggregation",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "loganalyzer.read"},
		func(ctx context.Context, _ *authz.Actor, in dto.ChartViewRequest) (any, error) {
			return uc.ChartView(ctx, in)
		})
}

// ---- loganalyzer.query.* ---------------------------------------------------

type laQueryUpsertInput struct {
	ID          uint64 `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Columns     string `json:"columns,omitempty"`
	Filters     string `json:"filters,omitempty"`
	Dataset     string `json:"dataset,omitempty"`
}

type laQueryListInput struct {
	Name  string `json:"name,omitempty"`
	Owner string `json:"owner,omitempty"`
	Page  int    `json:"page,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type laQueryIDInput struct {
	ID uint64 `json:"id"`
}

func registerLogAnalyzerQueries(m *Module) {
	uc := m.deps.LogAnalyzer.GetQueryUsecase()

	Add(m, &mcp.Tool{
		Name: "loganalyzer.query.create", Title: "Save analyzer query",
	}, Gate{Permission: "loganalyzer.write"},
		func(ctx context.Context, actor *authz.Actor, in laQueryUpsertInput) (any, error) {
			q := &ladomain.SavedQuery{
				Name: in.Name, Description: in.Description, Columns: in.Columns,
				Filters: in.Filters, Dataset: in.Dataset,
			}
			return uc.Create(ctx, q, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "loganalyzer.query.update", Title: "Update analyzer query",
	}, Gate{Permission: "loganalyzer.write"},
		func(ctx context.Context, actor *authz.Actor, in laQueryUpsertInput) (any, error) {
			q := &ladomain.SavedQuery{
				ID: in.ID, Name: in.Name, Description: in.Description, Columns: in.Columns,
				Filters: in.Filters, Dataset: in.Dataset,
			}
			return uc.Update(ctx, q, actor.Email)
		})

	Add(m, &mcp.Tool{
		Name: "loganalyzer.query.list", Title: "List saved queries",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "loganalyzer.read"},
		func(ctx context.Context, _ *authz.Actor, in laQueryListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.QueryFilter{
				Name: in.Name, Owner: in.Owner,
				Params: database.Params{Page: in.Page, Size: clampPageSize(in.Size)},
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "loganalyzer.query.get", Title: "Get saved query",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "loganalyzer.read"},
		func(ctx context.Context, _ *authz.Actor, in laQueryIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "loganalyzer.query.delete", Title: "Delete saved query",
	}, Gate{Permission: "loganalyzer.write"},
		func(ctx context.Context, _ *authz.Actor, in laQueryIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
