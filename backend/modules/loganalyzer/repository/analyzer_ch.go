package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

var datasets = map[string]store.Dataset{
	"logs":   eventstore.DatasetLogs,
	"alerts": eventstore.DatasetAlerts,
}

type chAnalyzerRepository struct {
	store  store.Reader
	events *eventstore.Store
}

func NewAnalyzerRepository(events *eventstore.Store) connectors.AnalyzerRepository {
	if events == nil {
		return nil
	}
	return &chAnalyzerRepository{store: events.GetReader(), events: events}
}

// rejectUnsupportedTextSearch turns a question the dataset cannot answer into a
// refusal the caller can read, rather than letting the driver fail mid-query.
func rejectUnsupportedTextSearch(dataset string, filters []common_models.FilterType) error {
	ds, ok := datasets[dataset]
	if !ok || eventstore.SupportsTextSearch(ds) {
		return nil
	}
	for _, f := range filters {
		if f.Operator == common_models.OpIsInFields || f.Operator == common_models.OpIsNotInFields {
			return domain.ErrNoTextSearch
		}
	}
	return nil
}

func scopeFor(ctx context.Context, dataset, dataType string) (store.Scope, error) {
	ds, ok := datasets[dataset]
	if !ok {
		return store.Scope{}, fmt.Errorf("%w: %s", domain.ErrUnknownDataset, dataset)
	}
	return store.Scope{
		Tenant:   authz.TenantIDFromContext(ctx),
		Dataset:  ds,
		DataType: dataType,
	}, nil
}

func (r *chAnalyzerRepository) TopValues(ctx context.Context, dataset, dataType, field string, filters []common_models.FilterType, top int) (*dto.TopValuesResponse, error) {
	scope, err := scopeFor(ctx, dataset, dataType)
	if err != nil {
		return nil, err
	}
	if err := rejectUnsupportedTextSearch(dataset, filters); err != nil {
		return nil, err
	}
	if err := common_models.ValidateField(field); err != nil {
		return nil, err
	}
	scope, filters = common_models.SplitTimeBounds(scope, filters)
	f, err := common_models.ToStoreFilters(filters)
	if err != nil {
		return nil, err
	}

	buckets, err := r.store.TopValues(ctx, scope, field, f, clampTop(top))
	if err != nil {
		return nil, err
	}
	total, err := r.store.Count(ctx, scope, f)
	if err != nil {
		return nil, err
	}

	out := &dto.TopValuesResponse{Total: total, Top: make([]dto.TopValue, 0, len(buckets))}
	for _, b := range buckets {
		v := dto.TopValue{Value: b.Key, Count: b.Count}
		if total > 0 {
			v.Percent = float64(b.Count) / float64(total) * 100
		}
		out.Top = append(out.Top, v)
	}
	return out, nil
}

func (r *chAnalyzerRepository) ChartView(ctx context.Context, req dto.ChartViewRequest) (*dto.ChartViewResponse, error) {
	scope, err := scopeFor(ctx, req.Dataset, req.DataType)
	if err != nil {
		return nil, err
	}
	if err := rejectUnsupportedTextSearch(req.Dataset, req.Filters); err != nil {
		return nil, err
	}
	scope, reqFilters := common_models.SplitTimeBounds(scope, req.Filters)
	f, err := common_models.ToStoreFilters(reqFilters)
	if err != nil {
		return nil, err
	}

	out := &dto.ChartViewResponse{Categories: []string{}, Values: []int64{}}

	if req.Interval != "" {
		interval, err := toInterval(req.Interval)
		if err != nil {
			return nil, err
		}
		points, err := r.store.Timeline(ctx, scope, f, interval)
		if err != nil {
			return nil, err
		}
		for _, p := range points {
			out.Categories = append(out.Categories, p.At.UTC().Format(time.RFC3339))
			out.Values = append(out.Values, p.Count)
		}
		return out, nil
	}

	buckets, err := r.store.TopValues(ctx, scope, req.Field, f, clampTop(req.Top))
	if err != nil {
		return nil, err
	}
	for _, b := range buckets {
		out.Categories = append(out.Categories, b.Key)
		out.Values = append(out.Values, b.Count)
	}
	return out, nil
}

func toInterval(s string) (store.Interval, error) {
	switch s {
	case "minute", "1m":
		return store.Interval{Duration: time.Minute}, nil
	case "hour", "1h":
		return store.Interval{Calendar: store.CalendarHour}, nil
	case "day", "1d":
		return store.Interval{Calendar: store.CalendarDay}, nil
	case "week", "1w":
		return store.Interval{Calendar: store.CalendarWeek}, nil
	}
	return store.Interval{}, fmt.Errorf("unknown interval: %s", s)
}

// Datasets is what a caller may explore. With an index store this was a table
// of patterns an operator maintained; the event store has two, and they are
// product knowledge rather than configuration.
func Datasets() []string { return []string{"logs", "alerts"} }

// Fields lists what can be filtered on and shown, which the caller needs to
// build anything. It comes from the store rather than a registry, so a field
// that exists is offered and one that does not is not.
func (r *chAnalyzerRepository) Fields(ctx context.Context, dataset string) ([]dto.Field, error) {
	scope, err := scopeFor(ctx, dataset, "")
	if err != nil {
		return nil, err
	}

	fields, err := r.store.DescribeFields(ctx, scope)
	if err != nil {
		return nil, err
	}

	out := make([]dto.Field, 0, len(fields))
	for _, f := range fields {
		out = append(out, dto.Field{
			Name:       f.Name,
			Type:       f.Type,
			Filterable: f.Filterable,
			Searchable: f.Searchable,
		})
	}
	return out, nil
}

const (
	defaultSearchSize = 50
	maxSearchSize     = 500
	defaultTopValues  = 10
	maxTopValues      = 1000
)

func clampTop(top int) int {
	if top <= 0 {
		return defaultTopValues
	}
	if top > maxTopValues {
		return maxTopValues
	}
	return top
}

// Search returns a page of documents for what the explorer has on screen.
//
// The window is part of the scope rather than a filter: the store narrows the
// table by time before anything else, and passing it as an ordinary predicate
// would forfeit that.
func (r *chAnalyzerRepository) Search(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error) {
	scope, err := scopeFor(ctx, req.Dataset, req.DataType)
	if err != nil {
		return nil, err
	}
	if req.From != nil {
		scope.From = *req.From
	}
	if req.To != nil {
		scope.To = *req.To
	}

	if err := rejectUnsupportedTextSearch(req.Dataset, req.Filters); err != nil {
		return nil, err
	}
	scope, reqFilters := common_models.SplitTimeBounds(scope, req.Filters)
	f, err := common_models.ToStoreFilters(reqFilters)
	if err != nil {
		return nil, err
	}

	size := req.Size
	if size <= 0 {
		size = defaultSearchSize
	}
	if size > maxSearchSize {
		size = maxSearchSize
	}
	page := req.Page
	if page < 0 {
		page = 0
	}

	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "@timestamp"
	}
	if err := common_models.ValidateField(sortBy); err != nil {
		return nil, err
	}
	order := store.Desc
	if req.Order == "asc" {
		order = store.Asc
	}

	docs, total, err := r.store.FetchPage(ctx, scope, f, store.Page{
		Offset: page * size,
		Limit:  size,
		SortBy: sortBy,
		Order:  order,
	})
	if err != nil {
		return nil, err
	}
	if docs == nil {
		docs = []json.RawMessage{}
	}
	if req.Dataset == "alerts" {
		if page == 0 {
			docs, total, err = r.includeRootsWithRecentEchoes(ctx, scope, f, docs, total, size-len(docs), sortBy, order)
			if err != nil {
				return nil, err
			}
		}
		docs = r.enrichEchoCounts(ctx, scope, docs)
	}
	return &dto.SearchResponse{Data: docs, Total: total}, nil
}

// includeRootsWithRecentEchoes closes ALR-F18: the root alert's own
// `@timestamp` is frozen at first occurrence, so a root that is still
// actively firing — but is itself older than the requested window — never
// passes the time filter above, and neither do its echoes (they only ever
// surface through GET /{id}/echoes, keyed by the root's id). A root that has
// a fresh echo inside [scope.From, scope.To] belongs in this result the same
// as one whose own timestamp falls in range.
//
// Scope: only applied to page 0 (room is clamped to what's left of the page
// size) — getting this fully correct across every page would mean sorting a
// result assembled from two separate queries by an arbitrary caller-chosen
// field, which the store's Filter model (AND-only, no OR/nesting, per
// store.Filter's doc comment) can't express as one query. Page 0 is what an
// analyst actually looks at after narrowing the time range, which is the
// scenario the finding describes; deeper pages keep the old behavior.
func (r *chAnalyzerRepository) includeRootsWithRecentEchoes(ctx context.Context, scope store.Scope, f []store.Filter, docs []json.RawMessage, total int64, room int, sortBy string, order store.SortOrder) ([]json.RawMessage, int64, error) {
	if room <= 0 || (scope.From.IsZero() && scope.To.IsZero()) {
		// No time window requested: every root's own timestamp already
		// passes it, so a child-only match couldn't add anything.
		return docs, total, nil
	}

	present := make(map[string]bool, len(docs))
	for _, raw := range docs {
		var m struct {
			ID string `json:"id"`
		}
		if json.Unmarshal(raw, &m) == nil && m.ID != "" {
			present[m.ID] = true
		}
	}

	buckets, err := r.store.TopValues(ctx, scope, "parentId",
		[]store.Filter{{Field: "parentId", Op: store.OpNotEq, Value: ""}}, maxActiveRootLookup)
	if err != nil {
		return nil, 0, err
	}

	missing := make([]string, 0, len(buckets))
	for _, b := range buckets {
		if b.Key != "" && b.Count > 0 && !present[b.Key] {
			missing = append(missing, b.Key)
		}
	}
	if len(missing) == 0 {
		return docs, total, nil
	}
	if len(missing) > room {
		missing = missing[:room]
	}

	unbounded := scope
	unbounded.From, unbounded.To = time.Time{}, time.Time{}
	rootFilters := append([]store.Filter{{Field: "id", Op: store.OpIn, Value: missing}}, f...)
	extra, _, err := r.store.FetchPage(ctx, unbounded, rootFilters,
		store.Page{Limit: len(missing), SortBy: sortBy, Order: order})
	if err != nil {
		return nil, 0, err
	}

	return append(docs, extra...), total + int64(len(extra)), nil
}

const maxActiveRootLookup = 10000

func (r *chAnalyzerRepository) enrichEchoCounts(ctx context.Context, scope store.Scope, docs []json.RawMessage) []json.RawMessage {
	if len(docs) == 0 {
		return docs
	}

	parsed := make([]map[string]any, len(docs))
	ids := make([]string, 0, len(docs))
	for i, raw := range docs {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		parsed[i] = m
		if id, ok := m["id"].(string); ok && id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return docs
	}

	// ponytail: one TopValues per page; fold into a single hand-written SELECT if the extra roundtrip ever shows up in traces.
	buckets, err := r.store.TopValues(ctx, scope, "parentId", []store.Filter{{Field: "parentId", Op: store.OpIn, Value: ids}}, len(ids))
	if err != nil {
		return docs
	}

	counts := make(map[string]int64, len(buckets))
	for _, b := range buckets {
		if b.Count > 0 {
			counts[b.Key] = b.Count
		}
	}

	for i, m := range parsed {
		if m == nil {
			continue
		}
		id, _ := m["id"].(string)
		if c, ok := counts[id]; ok {
			m["echoes"] = c
			if raw, err := json.Marshal(m); err == nil {
				docs[i] = raw
			}
		}
	}
	return docs
}

// DataTypes lists the kinds of record a dataset actually holds — o365,
// wineventlog and so on. It is what the explorer picks between: the dataset is
// the table, the data type is what an analyst thinks of as "which logs".
//
// Read from the data rather than from configuration, so a type that has stopped
// arriving stops being offered.
func (r *chAnalyzerRepository) DataTypes(ctx context.Context, dataset string) ([]string, error) {
	scope, err := scopeFor(ctx, dataset, "")
	if err != nil {
		return nil, err
	}
	buckets, err := r.store.TopValues(ctx, scope, dataTypeField, nil, maxDataTypes)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(buckets))
	for _, b := range buckets {
		if b.Key != "" {
			out = append(out, b.Key)
		}
	}
	return out, nil
}

const (
	dataTypeField = "dataType"
	maxDataTypes  = 200
)

func (r *chAnalyzerRepository) SearchSQL(ctx context.Context, sql string, page, size int) (*dto.SearchResponse, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		return nil, store.ErrNoTenant
	}
	if size <= 0 {
		size = defaultSearchSize
	}
	if size > maxSearchSize {
		size = maxSearchSize
	}

	scoped := common_models.ScopeSQL(sql, tenant, r.events.TableName(eventstore.DatasetLogs), r.events.TableName(eventstore.DatasetAlerts), page, size)

	rows, err := r.events.Query(ctx, scoped)
	if err != nil {
		return nil, err
	}

	return &dto.SearchResponse{Data: rows, Total: int64(len(rows))}, nil
}
