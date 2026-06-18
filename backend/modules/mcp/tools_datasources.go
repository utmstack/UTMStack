package mcp

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/modules/datasources/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

func registerDatasources(m *Module) {
	registerDatasourceDatasources(m)
	registerDatasourceGroups(m)
}

// ---- datasources.* ---------------------------------------------------------

type dsListInput struct {
	PageNumber  int    `json:"page_number,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	SearchQuery string `json:"search_query,omitempty" jsonschema:"DSL e.g. name.contains.foo&dataType.equals.linux"`
	SortBy      string `json:"sort_by,omitempty"`
}

type dsIDInput struct {
	ID uint64 `json:"id"`
}

type dsUpdateGroupInput struct {
	IDs     []uint64 `json:"ids"`
	GroupID *uint64  `json:"group_id,omitempty" jsonschema:"null/omitted to unset"`
}

type dsUpdateLabelsInput struct {
	ID     uint64 `json:"id"`
	Labels string `json:"labels" jsonschema:"Comma-separated labels"`
}

func registerDatasourceDatasources(m *Module) {
	uc := m.deps.Datasources.GetDatasourceUsecase()
	lic := m.deps.Datasources.LicenseCap()

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
		Name: "datasources.update_group", Title: "Move datasources between groups",
		Annotations: &mcp.ToolAnnotations{IdempotentHint: true},
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsUpdateGroupInput) (any, error) {
			if err := uc.UpdateGroup(ctx, dto.UpdateGroupRequest{IDs: in.IDs, GroupID: in.GroupID}); err != nil {
				return nil, err
			}
			return map[string]any{"updated": len(in.IDs)}, nil
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
		Name: "datasources.usage", Title: "Datasource usage vs license cap",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "datasources.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			count, err := uc.Count(ctx)
			if err != nil {
				return nil, err
			}
			var limit int64
			unlimited := true
			if lic != nil {
				limit, unlimited = lic.DatasourceCap()
			}
			return dto.UsageResponse{
				Count: count, Limit: limit, Unlimited: unlimited,
				OverLimit: !unlimited && count > limit,
			}, nil
		})
}

// ---- datasource_groups.* ---------------------------------------------------

type dsGroupUpsertInput struct {
	ID               uint64 `json:"id,omitempty"`
	GroupName        string `json:"group_name"`
	GroupDescription string `json:"group_description,omitempty"`
}

func registerDatasourceGroups(m *Module) {
	uc := m.deps.Datasources.GetAssetGroupUsecase()

	Add(m, &mcp.Tool{
		Name: "datasource_groups.create", Title: "Create datasource group",
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsGroupUpsertInput) (any, error) {
			return uc.Create(ctx, &domain.UtmAssetGroup{
				GroupName: in.GroupName, GroupDescription: in.GroupDescription,
			})
		})

	Add(m, &mcp.Tool{
		Name: "datasource_groups.update", Title: "Update datasource group",
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsGroupUpsertInput) (any, error) {
			return uc.Update(ctx, &domain.UtmAssetGroup{
				ID: in.ID, GroupName: in.GroupName, GroupDescription: in.GroupDescription,
			})
		})

	Add(m, &mcp.Tool{
		Name: "datasource_groups.list", Title: "List datasource groups",
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
		Name: "datasource_groups.get", Title: "Get datasource group",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "datasources.read"},
		func(ctx context.Context, _ *authz.Actor, in dsIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "datasource_groups.delete", Title: "Delete datasource group",
	}, Gate{Permission: "datasources.write"},
		func(ctx context.Context, _ *authz.Actor, in dsIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
}
