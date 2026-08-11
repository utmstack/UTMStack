package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

// registerStore mounts the tools that speak to the event store (ClickHouse).
// The permission stays "opensearch.read" so existing roles keep working; only
// the tool names and shapes moved.
func registerStore(m *Module) {
	events := m.deps.Events
	if events == nil {
		// No ClickHouse configured; store.* tools are unusable. Skip
		// registration so /mcp/health does not advertise tools that always
		// fail.
		return
	}

	registerStoreQueries(m, events)
	registerStoreIntrospection(m, events)
}

var storeDatasets = map[string]store.Dataset{
	"logs":       eventstore.DatasetLogs,
	"alerts":     eventstore.DatasetAlerts,
	"statistics": eventstore.DatasetStats,
}

func resolveDataset(name string) (store.Dataset, error) {
	if name == "" {
		return "", fmt.Errorf("dataset is required (logs | alerts | statistics)")
	}
	ds, ok := storeDatasets[name]
	if !ok {
		return "", fmt.Errorf("unknown dataset %q (must be logs | alerts | statistics)", name)
	}
	return ds, nil
}

func storeScope(ctx context.Context, dataset, dataType string) (store.Scope, error) {
	ds, err := resolveDataset(dataset)
	if err != nil {
		return store.Scope{}, err
	}
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" && tenancy.ReadsAllTenants(ctx) {
		tenant = store.AllTenants
	}
	if tenant == "" {
		return store.Scope{}, fmt.Errorf("tenant required")
	}
	return store.Scope{Tenant: tenant, Dataset: ds, DataType: dataType}, nil
}

// ---- store.search / store.count / store.property_values --------------------

type storeSearchInput struct {
	Dataset   string                     `json:"dataset" jsonschema:"logs | alerts | statistics"`
	DataType  string                     `json:"data_type,omitempty" jsonschema:"e.g. wineventlog, o365 — filters by the dataType column"`
	Filters   []common_models.FilterType `json:"filters,omitempty" jsonschema:"FilterType DSL predicates (AND'd)"`
	Page      int                        `json:"page,omitempty" jsonschema:"1-based page"`
	Size      int                        `json:"size,omitempty"`
	SortBy    string                     `json:"sort_by,omitempty"`
	SortOrder string                     `json:"sort_order,omitempty" jsonschema:"asc | desc"`
}

type storeCountInput struct {
	Dataset  string                     `json:"dataset"`
	DataType string                     `json:"data_type,omitempty"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
}

type storePropertyValuesInput struct {
	Dataset  string                     `json:"dataset"`
	DataType string                     `json:"data_type,omitempty"`
	Field    string                     `json:"field" jsonschema:"Field name to bucket by"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
	Top      int                        `json:"top,omitempty"`
}

type storeSearchSQLInput struct {
	Query string `json:"query" jsonschema:"SELECT-only ClickHouse SQL. The tenant scope is injected as WITH logs/alerts CTEs; do not read the physical tables directly."`
	Page  int    `json:"page,omitempty"`
	Size  int    `json:"size,omitempty"`
}

type storeSearchCSVInput struct {
	Dataset  string                     `json:"dataset"`
	DataType string                     `json:"data_type,omitempty"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
	Columns  []string                   `json:"columns" jsonschema:"Field names to project"`
	Top      int                        `json:"top,omitempty"`
}

func registerStoreQueries(m *Module, events *eventstore.Store) {
	Add(m, &mcp.Tool{
		Name: "store.search", Title: "Structured search",
		Description: "Run a structured search over a store dataset (logs | alerts | statistics) using the FilterType DSL. See mcp://utmstack/docs/filter-operators.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storeSearchInput) (any, error) {
			scope, err := storeScope(ctx, in.Dataset, in.DataType)
			if err != nil {
				return nil, err
			}
			scope, rest := common_models.SplitTimeBounds(scope, in.Filters)
			f, err := common_models.ToStoreFilters(rest)
			if err != nil {
				return nil, err
			}

			page, size := in.Page, in.Size
			if page < 1 {
				page = 1
			}
			size = clampPageSize(size)
			order := store.Desc
			if in.SortOrder == "asc" {
				order = store.Asc
			}
			sortBy := in.SortBy
			if sortBy == "" {
				sortBy = "@timestamp"
			}

			rows, total, err := events.FetchPage(ctx, scope, f, store.Page{
				Offset: (page - 1) * size,
				Limit:  size,
				SortBy: sortBy,
				Order:  order,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"hits": rowsToMaps(rows), "total": total}, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.count", Title: "Count matching docs",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storeCountInput) (any, error) {
			scope, err := storeScope(ctx, in.Dataset, in.DataType)
			if err != nil {
				return nil, err
			}
			scope, rest := common_models.SplitTimeBounds(scope, in.Filters)
			f, err := common_models.ToStoreFilters(rest)
			if err != nil {
				return nil, err
			}
			total, err := events.Count(ctx, scope, f)
			if err != nil {
				return nil, err
			}
			return map[string]any{"total": total, "hasResults": total > 0}, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.property_values", Title: "Top values for a field",
		Description: "Return the top-N distinct values of a field with their document counts.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storePropertyValuesInput) (any, error) {
			if in.Field == "" {
				return nil, fmt.Errorf("field is required")
			}
			scope, err := storeScope(ctx, in.Dataset, in.DataType)
			if err != nil {
				return nil, err
			}
			scope, rest := common_models.SplitTimeBounds(scope, in.Filters)
			f, err := common_models.ToStoreFilters(rest)
			if err != nil {
				return nil, err
			}
			n := in.Top
			if n <= 0 {
				n = 100
			}
			buckets, err := events.TopValues(ctx, scope, in.Field, f, n)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]any, 0, len(buckets))
			for _, b := range buckets {
				out = append(out, map[string]any{"value": b.Key, "count": b.Count})
			}
			return out, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.search_sql", Title: "SQL search",
		Description: "Run a read-only ClickHouse SQL query. The FROM clause must reference the logs / alerts CTEs (already tenant-scoped) rather than the physical tables.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storeSearchSQLInput) (any, error) {
			if in.Query == "" {
				return nil, fmt.Errorf("query is required")
			}
			tenant := authz.TenantIDFromContext(ctx)
			if tenant == "" {
				return nil, fmt.Errorf("tenant required")
			}
			page, size := in.Page, in.Size
			if page < 1 {
				page = 1
			}
			size = clampPageSize(size)

			scoped := storeScopedSQL(in.Query,
				events.TableName(eventstore.DatasetLogs),
				events.TableName(eventstore.DatasetAlerts),
				page, size)

			qctx := clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
				"readonly":             1,
				"max_execution_time":   30,
				"max_result_rows":      uint64(size),
				"result_overflow_mode": "break",
			}))

			rows, err := events.Conn.Query(qctx, scoped, tenant, tenant)
			if err != nil {
				return nil, err
			}
			defer func() { _ = rows.Close() }()

			hits := make([]map[string]any, 0)
			cols := rows.Columns()
			types := rows.ColumnTypes()
			for rows.Next() {
				ptrs := make([]any, len(cols))
				for i := range ptrs {
					ptrs[i] = reflect.New(types[i].ScanType()).Interface()
				}
				if err := rows.Scan(ptrs...); err != nil {
					return nil, err
				}
				doc := make(map[string]any, len(cols))
				for i, c := range cols {
					doc[c] = reflect.ValueOf(ptrs[i]).Elem().Interface()
				}
				hits = append(hits, doc)
			}
			if err := rows.Err(); err != nil {
				return nil, err
			}
			return map[string]any{"hits": hits, "total": int64(len(hits))}, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.search_csv", Title: "Search and project rows to columns",
		Description: "Same as store.search but returns rows projected to the requested column list — cheaper for the model to consume when only a few fields matter.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storeSearchCSVInput) (any, error) {
			if len(in.Columns) == 0 {
				return nil, fmt.Errorf("columns is required")
			}
			scope, err := storeScope(ctx, in.Dataset, in.DataType)
			if err != nil {
				return nil, err
			}
			scope, rest := common_models.SplitTimeBounds(scope, in.Filters)
			f, err := common_models.ToStoreFilters(rest)
			if err != nil {
				return nil, err
			}
			top := in.Top
			if top <= 0 {
				top = 500
			}
			rows, _, err := events.FetchPage(ctx, scope, f, store.Page{
				Offset: 0, Limit: clampPageSize(top),
				SortBy: "@timestamp", Order: store.Desc,
			})
			if err != nil {
				return nil, err
			}
			out := make([]map[string]any, 0, len(rows))
			for _, raw := range rows {
				var doc map[string]any
				if err := json.Unmarshal(raw, &doc); err != nil {
					continue
				}
				projected := make(map[string]any, len(in.Columns))
				for _, c := range in.Columns {
					projected[c] = doc[c]
				}
				out = append(out, projected)
			}
			return out, nil
		})
}

// ---- store.datasets / store.dataset.fields ---------------------------------

type storeDatasetFieldsInput struct {
	Dataset string `json:"dataset"`
}

func registerStoreIntrospection(m *Module, events *eventstore.Store) {
	Add(m, &mcp.Tool{
		Name: "store.datasets", Title: "List datasets",
		Description: "Return the datasets the store exposes (logs, alerts, statistics). These are the values callers pass as `dataset` on the other store.* tools.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(_ context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return []string{"logs", "alerts", "statistics"}, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.dataset.fields", Title: "List fields for a dataset",
		Description: "Describe the queryable fields of a dataset (name, type, filterable, searchable).",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(ctx context.Context, _ *authz.Actor, in storeDatasetFieldsInput) (any, error) {
			scope, err := storeScope(ctx, in.Dataset, "")
			if err != nil {
				return nil, err
			}
			fields, err := events.DescribeFields(ctx, scope)
			if err != nil {
				return nil, err
			}
			out := make([]map[string]any, 0, len(fields))
			for _, f := range fields {
				out = append(out, map[string]any{
					"name":       f.Name,
					"type":       f.Type,
					"filterable": f.Filterable,
					"searchable": f.Searchable,
				})
			}
			return out, nil
		})
}

// ---- helpers ---------------------------------------------------------------

func rowsToMaps(rows []json.RawMessage) []map[string]any {
	out := make([]map[string]any, 0, len(rows))
	for _, raw := range rows {
		var doc map[string]any
		if err := json.Unmarshal(raw, &doc); err != nil {
			continue
		}
		out = append(out, doc)
	}
	return out
}

// storeScopedSQL wraps a caller-supplied SELECT with logs/alerts CTEs already
// scoped to the current tenant — the same trick loganalyzer uses to keep the
// SQL surface safe without a parser.
func storeScopedSQL(query, logsTable, alertsTable string, page, size int) string {
	if page < 1 {
		page = 1
	}
	q := trimTrailingSemicolons(query)
	return fmt.Sprintf(
		"WITH logs AS (SELECT * FROM %s WHERE tenantId = ?), alerts AS (SELECT * FROM %s WHERE tenantId = ?) "+
			"SELECT * FROM (%s) LIMIT %d OFFSET %d",
		logsTable, alertsTable, q, size, (page-1)*size,
	)
}

func trimTrailingSemicolons(q string) string {
	for len(q) > 0 && (q[len(q)-1] == ';' || q[len(q)-1] == ' ' || q[len(q)-1] == '\n' || q[len(q)-1] == '\t') {
		q = q[:len(q)-1]
	}
	return q
}
