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
	DataType  string                     `json:"data_type,omitempty" jsonschema:"filter by dataType (e.g. wineventlog, o365)"`
	Filters   []common_models.FilterType `json:"filters,omitempty" jsonschema:"AND'd FilterType predicates"`
	Page      int                        `json:"page,omitempty" jsonschema:"1-based; default 1"`
	Size      int                        `json:"size,omitempty" jsonschema:"rows per page; default 50, max 500"`
	SortBy    string                     `json:"sort_by,omitempty" jsonschema:"field name; default @timestamp"`
	SortOrder string                     `json:"sort_order,omitempty" jsonschema:"asc | desc; default desc"`
}

type storeCountInput struct {
	Dataset  string                     `json:"dataset" jsonschema:"logs | alerts | statistics"`
	DataType string                     `json:"data_type,omitempty" jsonschema:"filter by dataType"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
}

type storePropertyValuesInput struct {
	Dataset  string                     `json:"dataset" jsonschema:"logs | alerts | statistics"`
	DataType string                     `json:"data_type,omitempty" jsonschema:"filter by dataType"`
	Field    string                     `json:"field" jsonschema:"field to group by"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
	Top      int                        `json:"top,omitempty" jsonschema:"default 100"`
}

type storeSearchSQLInput struct {
	Query string `json:"query" jsonschema:"SELECT-only. Use the tenant-scoped CTEs 'logs' and 'alerts' as FROM tables."`
	Page  int    `json:"page,omitempty" jsonschema:"1-based; default 1"`
	Size  int    `json:"size,omitempty" jsonschema:"default 50, max 500"`
}

type storeSearchCSVInput struct {
	Dataset  string                     `json:"dataset" jsonschema:"logs | alerts | statistics"`
	DataType string                     `json:"data_type,omitempty" jsonschema:"filter by dataType"`
	Filters  []common_models.FilterType `json:"filters,omitempty"`
	Columns  []string                   `json:"columns" jsonschema:"fields to project"`
	Top      int                        `json:"top,omitempty" jsonschema:"default 500, max 500"`
}

func registerStoreQueries(m *Module, events *eventstore.Store) {
	Add(m, &mcp.Tool{
		Name: "store.search", Title: "Search a dataset",
		Description: "Return paged hits from logs | alerts | statistics matching the FilterType predicates. See mcp://utmstack/docs/filter-operators.",
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
		Description: "Return the total number of docs in the dataset that match the filters.",
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
		Description: "Top-N distinct values of a field with their counts.",
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
		Name: "store.search_sql", Title: "SQL query",
		Description: "Run a SELECT-only ClickHouse query against the tenant-scoped 'logs' and 'alerts' CTEs.",
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
		Name: "store.search_csv", Title: "Search, project to columns",
		Description: "Like store.search but each row is trimmed to the requested columns.",
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
	Dataset string `json:"dataset" jsonschema:"logs | alerts | statistics"`
}

func registerStoreIntrospection(m *Module, events *eventstore.Store) {
	Add(m, &mcp.Tool{
		Name: "store.datasets", Title: "List datasets",
		Description: "The dataset names accepted by every store.* tool: logs, alerts, statistics.",
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}, Gate{Permission: "opensearch.read"},
		func(_ context.Context, _ *authz.Actor, _ struct{}) (any, error) {
			return []string{"logs", "alerts", "statistics"}, nil
		})

	Add(m, &mcp.Tool{
		Name: "store.dataset.fields", Title: "Dataset fields",
		Description: "Queryable fields of a dataset: name, type, filterable, searchable.",
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
