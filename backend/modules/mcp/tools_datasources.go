package mcp

import (
	"context"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type dsListInput struct {
	PageNumber  int    `json:"page_number,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	SearchQuery string `json:"search_query,omitempty" jsonschema:"DSL e.g. name.contains.foo&dataType.equals.linux"`
	SortBy      string `json:"sort_by,omitempty"`
}

type dsIDInput struct {
	ID uuid.UUID `json:"id"`
}

type dsUpdateLabelsInput struct {
	ID     uuid.UUID `json:"id"`
	Labels string    `json:"labels" jsonschema:"Comma-separated labels"`
}

func registerDatasources(m *Module) {
	uc := m.deps.Datasources.GetDatasourceUsecase()

	Add(m, &mcp.Tool{
		Name: "datasources.list", Title: "List datasources",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "datasources.read"},
		func(ctx context.Context, _ *authz.Actor, in dsListInput) (any, error) {
			req := common_models.ListRequest{
				PageNumber: in.PageNumber, PageSize: clampPageSize(in.PageSize),
				SearchQuery: in.SearchQuery, SortBy: in.SortBy,
			}
			return uc.List(ctx, req)
		})

	Add(m, &mcp.Tool{
		Name: "datasources.get", Title: "Get datasource",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "datasources.read"},
		func(ctx context.Context, _ *authz.Actor, in dsIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "datasources.update_labels", Title: "Set datasource labels",
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsUpdateLabelsInput) (any, error) {
			if err := uc.UpdateLabels(ctx, dto.UpdateLabelsRequest{ID: in.ID, Labels: in.Labels}); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "labels": in.Labels}, nil
		})

	Add(m, &mcp.Tool{
		Name: "datasources.delete", Title: "Delete datasource",
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})

	Add(m, &mcp.Tool{
		Name: "datasources.count", Title: "Number of configured datasources",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "datasources.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			count, err := uc.Count(ctx)
			if err != nil {
				return nil, err
			}
			return dto.CountResponse{Count: count}, nil
		})
}
