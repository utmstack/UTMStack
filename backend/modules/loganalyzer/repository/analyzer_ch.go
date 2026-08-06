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

type Reader interface {
	DescribeFields(ctx context.Context, s store.Scope) ([]store.Field, error)
	FetchPage(ctx context.Context, s store.Scope, filters []store.Filter, page store.Page) ([]json.RawMessage, int64, error)
	TopValues(ctx context.Context, s store.Scope, field string, filters []store.Filter, n int) ([]store.Bucket, error)
	Timeline(ctx context.Context, s store.Scope, filters []store.Filter, interval store.Interval) ([]store.Point, error)
	Count(ctx context.Context, s store.Scope, filters []store.Filter) (int64, error)
}

type chAnalyzerRepository struct{ store Reader }

func NewAnalyzerRepository(r Reader) connectors.AnalyzerRepository {
	if r == nil {
		return nil
	}
	return &chAnalyzerRepository{store: r}
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
	f, err := common_models.ToStoreFilters(filters)
	if err != nil {
		return nil, err
	}

	if top <= 0 {
		top = 10
	}
	buckets, err := r.store.TopValues(ctx, scope, field, f, top)
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
	f, err := common_models.ToStoreFilters(req.Filters)
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

	top := req.Top
	if top <= 0 {
		top = 10
	}
	buckets, err := r.store.TopValues(ctx, scope, req.Field, f, top)
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
)

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

	f, err := common_models.ToStoreFilters(req.Filters)
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
	return &dto.SearchResponse{Data: docs, Total: total}, nil
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
