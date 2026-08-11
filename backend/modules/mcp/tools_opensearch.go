package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

func registerOpenSearch(m *Module) {
	registerOSGateway(m)
	registerOSIndexPatterns(m)
	registerOSPolicy(m)
}

// ---- opensearch search / index gateway -------------------------------------

type osSearchInput struct {
	IndexPattern    string                     `json:"index_pattern" jsonschema:"Required. Resolved index pattern with syntax v11-[name]-* (e.g. v11-alert-*, v11-log-*)."`
	Filters         []common_models.FilterType `json:"filters,omitempty" jsonschema:"FilterType DSL conditions"`
	Top             int                        `json:"top,omitempty" jsonschema:"Max total hits to return (default 500)"`
	Page            int                        `json:"page,omitempty" jsonschema:"1-based page"`
	Size            int                        `json:"size,omitempty"`
	SortBy          string                     `json:"sort_by,omitempty"`
	SortOrder       string                     `json:"sort_order,omitempty" jsonschema:"asc | desc"`
	IncludeChildren bool                       `json:"include_children,omitempty"`
}

type osGenericSearchInput struct {
	Index   string                     `json:"index"`
	Filters []common_models.FilterType `json:"filters,omitempty"`
	Top     int                        `json:"top,omitempty"`
	Page    int                        `json:"page,omitempty"`
	Size    int                        `json:"size,omitempty"`
}

type osCountInput struct {
	IndexPattern string                     `json:"index_pattern"`
	Filters      []common_models.FilterType `json:"filters,omitempty"`
}

type osPropertyValuesInput struct {
	IndexPattern string `json:"index_pattern"`
	Keyword      string `json:"keyword"`
}

type osPropertyValuesWithCountInput struct {
	Index        string                     `json:"index"`
	Field        string                     `json:"field"`
	Filters      []common_models.FilterType `json:"filters,omitempty"`
	Top          int                        `json:"top,omitempty"`
	OrderByCount bool                       `json:"order_by_count,omitempty"`
	SortAsc      bool                       `json:"sort_asc,omitempty"`
}

type osIndexPropertiesInput struct {
	Pattern string `json:"pattern"`
}

type osIndexListInput struct {
	Pattern       string `json:"pattern,omitempty"`
	IncludeSystem bool   `json:"include_system,omitempty"`
	Page          int    `json:"page,omitempty"`
	Size          int    `json:"size,omitempty"`
}

type osIndexDeleteInput struct {
	Indices []string `json:"indices" jsonschema:"Exact index names to delete"`
}

type osSearchSQLInput struct {
	Query     string `json:"query" jsonschema:"OpenSearch SQL statement; SELECT-only, sanitized at runtime"`
	FetchSize *int   `json:"fetch_size,omitempty"`
	Page      int    `json:"page,omitempty"`
	Size      int    `json:"size,omitempty"`
}

type osSearchCSVInput struct {
	IndexPattern string                     `json:"index_pattern"`
	Filters      []common_models.FilterType `json:"filters,omitempty"`
	Top          int                        `json:"top,omitempty"`
	Columns      []dto.DataColumn           `json:"columns"`
}

func registerOSGateway(m *Module) {
	gw := m.deps.OpenSearch.Gateway()

	Add(m, &mcp.Tool{
		Name: "opensearch.search", Title: "Structured search",
		Description: "Run a structured search against the given index pattern using the FilterType DSL. See mcp://utmstack/docs/filter-operators for operator semantics.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osSearchInput) (any, error) {
			if in.IndexPattern == "" {
				return nil, fmt.Errorf("index_pattern is required")
			}
			page := in.Page
			if page <= 0 {
				page = 1
			}
			hits, total, err := gw.Search(ctx, in.Filters, in.Top, in.IndexPattern, in.IncludeChildren, page, clampPageSize(in.Size), in.SortBy, in.SortOrder)
			if err != nil {
				return nil, err
			}
			return map[string]any{"hits": hits, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.generic_search", Title: "Search any index",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osGenericSearchInput) (any, error) {
			if in.Index == "" {
				return nil, fmt.Errorf("index is required")
			}
			page := in.Page
			if page <= 0 {
				page = 1
			}
			hits, total, err := gw.GenericSearch(ctx, in.Index, in.Filters, in.Top, page, clampPageSize(in.Size))
			if err != nil {
				return nil, err
			}
			return map[string]any{"hits": hits, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.count", Title: "Count matching docs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osCountInput) (any, error) {
			hasResults, err := gw.Count(ctx, in.Filters, in.IndexPattern)
			if err != nil {
				return nil, err
			}
			return map[string]any{"hasResults": hasResults}, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.search_sql", Title: "SQL search",
		Description: "Run a sanitized SQL query against the OpenSearch SQL plugin. SELECT-only.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osSearchSQLInput) (any, error) {
			page := in.Page
			if page <= 0 {
				page = 1
			}
			hits, total, err := gw.SearchSQL(ctx, dto.SqlSearchDto{Query: in.Query, FetchSize: in.FetchSize}, page, clampPageSize(in.Size))
			if err != nil {
				return nil, err
			}
			return map[string]any{"hits": hits, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.search_csv", Title: "Project search results to CSV columns",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osSearchCSVInput) (any, error) {
			if len(in.Columns) == 0 {
				return nil, fmt.Errorf("columns is required")
			}
			rows, err := gw.SearchCSV(ctx, in.Filters, in.Top, in.IndexPattern, in.Columns)
			if err != nil {
				return nil, err
			}
			return rows, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.property_values", Title: "Distinct values for a field",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osPropertyValuesInput) (any, error) {
			return gw.PropertyValues(ctx, dto.PropertyValuesRequest{
				Keyword: in.Keyword, IndexPattern: in.IndexPattern,
			})
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.property_values_with_count", Title: "Field values with counts",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osPropertyValuesWithCountInput) (any, error) {
			return gw.PropertyValuesWithCount(ctx, dto.PropertyValuesWithCountRequest{
				Filters: in.Filters, Field: in.Field, Index: in.Index,
				Top: in.Top, OrderByCount: in.OrderByCount, SortAsc: in.SortAsc,
			})
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.index.properties", Title: "List index fields",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osIndexPropertiesInput) (any, error) {
			return gw.IndexProperties(ctx, in.Pattern)
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.index.list", Title: "List indices matching a pattern",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osIndexListInput) (any, error) {
			page := in.Page
			if page <= 0 {
				page = 1
			}
			items, total, err := gw.IndexAll(ctx, in.Pattern, in.IncludeSystem, page, clampPageSize(in.Size))
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	/* temporarily disabled — index/retention management not exposed to the agent for now
	Add(m, &mcp.Tool{
		Name: "opensearch.index.delete", Title: "Delete indices",
		Description: "Permanently delete one or more indices. Destructive.",
	}, Gate{Permission: "opensearch.write"},
		func(ctx context.Context, _ *authz.Actor, in osIndexDeleteInput) (any, error) {
			if len(in.Indices) == 0 {
				return nil, fmt.Errorf("indices is required")
			}
			if err := gw.DeleteIndex(ctx, in.Indices); err != nil {
				return nil, err
			}
			return map[string]any{"deleted": len(in.Indices)}, nil
		})
	*/

	Add(m, &mcp.Tool{
		Name: "opensearch.cluster_status", Title: "Cluster status",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return gw.ClusterStatus(ctx)
		})
}

// ---- opensearch.index_pattern.* --------------------------------------------

type osPatternCreateInput struct {
	Pattern       string  `json:"pattern"`
	PatternModule *string `json:"pattern_module,omitempty"`
	PatternSystem *bool   `json:"pattern_system,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

type osPatternUpdateInput struct {
	ID            int64   `json:"id"`
	Pattern       string  `json:"pattern"`
	PatternModule *string `json:"pattern_module,omitempty"`
	PatternSystem *bool   `json:"pattern_system,omitempty"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

type osPatternListInput struct {
	PatternContains *string `json:"pattern_contains,omitempty"`
	ModuleEquals    *string `json:"module_equals,omitempty"`
	IsActive        *bool   `json:"is_active,omitempty"`
	Page            int     `json:"page,omitempty"`
	Size            int     `json:"size,omitempty"`
	Sort            string  `json:"sort,omitempty"`
}

type osPatternIDInput struct {
	ID int64 `json:"id"`
}

func registerOSIndexPatterns(m *Module) {
	uc := m.deps.OpenSearch.GetIndexPatternUsecase()

	/* temporarily disabled — index/retention management not exposed to the agent for now
	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.create", Title: "Create index pattern",
	}, Gate{Permission: "opensearch.write"},
		func(ctx context.Context, _ *authz.Actor, in osPatternCreateInput) (any, error) {
			return uc.Create(ctx, dto.CreateIndexPatternRequest{
				Pattern: in.Pattern, PatternModule: in.PatternModule,
				PatternSystem: in.PatternSystem, IsActive: in.IsActive,
			})
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.update", Title: "Update index pattern",
	}, Gate{Permission: "opensearch.write"},
		func(ctx context.Context, _ *authz.Actor, in osPatternUpdateInput) (any, error) {
			return uc.Update(ctx, dto.UpdateIndexPatternRequest{
				ID: in.ID, Pattern: in.Pattern, PatternModule: in.PatternModule,
				PatternSystem: in.PatternSystem, IsActive: in.IsActive,
			})
		})
	*/

	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.list", Title: "List index patterns",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osPatternListInput) (any, error) {
			items, total, err := uc.List(ctx, dto.IndexPatternFilters{
				PatternContains: in.PatternContains, ModuleEquals: in.ModuleEquals, IsActive: in.IsActive,
				Sort: in.Sort,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.get", Title: "Get index pattern",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osPatternIDInput) (any, error) {
			return uc.GetByID(ctx, in.ID)
		})

	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.fields", Title: "Fields for index pattern",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in osPatternListInput) (any, error) {
			items, total, err := uc.Fields(ctx, dto.IndexPatternFilters{
				PatternContains: in.PatternContains, ModuleEquals: in.ModuleEquals, IsActive: in.IsActive,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"items": items, "total": total}, nil
		})

	/* temporarily disabled — index/retention management not exposed to the agent for now
	Add(m, &mcp.Tool{
		Name: "opensearch.index_pattern.delete", Title: "Delete index pattern",
	}, Gate{Permission: "opensearch.write"},
		func(ctx context.Context, _ *authz.Actor, in osPatternIDInput) (any, error) {
			if err := uc.Delete(ctx, in.ID); err != nil {
				return nil, err
			}
			return map[string]any{"id": in.ID, "deleted": true}, nil
		})
	*/
}

// ---- opensearch.policy.* ---------------------------------------------------

type osPolicyUpdateInput struct {
	SnapshotActive bool   `json:"snapshot_active"`
	DeleteAfter    string `json:"delete_after"`
}

func registerOSPolicy(m *Module) {
	uc := m.deps.OpenSearch.GetPolicyUsecase()
	if uc == nil {
		// OpenSearch not configured; policy tools are unusable. Skip registration
		// so /mcp/health doesn't advertise tools that always fail.
		return
	}

	Add(m, &mcp.Tool{
		Name: "opensearch.policy.get", Title: "Get ISM policy",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return uc.GetPolicy(ctx)
		})

	/* temporarily disabled — index/retention management not exposed to the agent for now
	Add(m, &mcp.Tool{
		Name: "opensearch.policy.update", Title: "Update ISM policy",
	}, Gate{Permission: "opensearch.write"},
		func(ctx context.Context, _ *authz.Actor, in osPolicyUpdateInput) (any, error) {
			return uc.UpdatePolicy(ctx, domain.PolicySettings{
				SnapshotActive: in.SnapshotActive, DeleteAfter: in.DeleteAfter,
			})
		})
	*/
}
