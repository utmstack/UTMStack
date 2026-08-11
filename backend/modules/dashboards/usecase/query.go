package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"errors"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	queryDefaultLimit = 50
	queryMaxLimit     = 10000
)

type Reader interface {
	Count(ctx context.Context, s store.Scope, filters []store.Filter) (int64, error)
	TopValues(ctx context.Context, s store.Scope, field string, filters []store.Filter, n int) ([]store.Bucket, error)
	Timeline(ctx context.Context, s store.Scope, filters []store.Filter, interval store.Interval) ([]store.Point, error)
	TimelineByField(ctx context.Context, s store.Scope, field string, filters []store.Filter, interval store.Interval, n int) ([]store.Series, error)
	FetchPage(ctx context.Context, s store.Scope, filters []store.Filter, page store.Page) ([]json.RawMessage, int64, error)
}

type QueryService struct{ store Reader }

func NewQueryService(r Reader) *QueryService { return &QueryService{store: r} }

// Result is one answer, in the shape the chart that asked for it reads. The
// store's own types are remapped rather than returned: they carry no JSON tags,
// so serving them would put Go field names on the wire.
type Result struct {
	Total   *int64            `json:"total,omitempty"`
	Buckets []Bucket          `json:"buckets,omitempty"`
	Points  []Point           `json:"points,omitempty"`
	Series  []Series          `json:"series,omitempty"`
	Rows    []json.RawMessage `json:"rows,omitempty"`
}

// Bucket is one value of the field a chart breaks down by.
type Bucket struct {
	Key   string `json:"key"`
	Count int64  `json:"count"`
}

// Point is one time bucket.
type Point struct {
	At    time.Time `json:"at"`
	Count int64     `json:"count"`
}

// Series is one line: a value of the split field and its points.
type Series struct {
	Key    string  `json:"key"`
	Points []Point `json:"points"`
}

func toBuckets(in []store.Bucket) []Bucket {
	out := make([]Bucket, 0, len(in))
	for _, b := range in {
		out = append(out, Bucket{Key: b.Key, Count: b.Count})
	}
	return out
}

func toPoints(in []store.Point) []Point {
	out := make([]Point, 0, len(in))
	for _, p := range in {
		out = append(out, Point{At: p.At, Count: p.Count})
	}
	return out
}

func toSeries(in []store.Series) []Series {
	out := make([]Series, 0, len(in))
	for _, s := range in {
		out = append(out, Series{Key: s.Key, Points: toPoints(s.Points)})
	}
	return out
}

func (s *QueryService) Run(ctx context.Context, spec domain.Spec) (*Result, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	tenant, err := scopeTenant(ctx)
	if err != nil {
		return nil, err
	}
	scope := store.Scope{
		Tenant:   tenant,
		Dataset:  store.Dataset(spec.Dataset),
		DataType: spec.DataType,
	}
	if spec.From != nil {
		scope.From = *spec.From
	}
	if spec.To != nil {
		scope.To = *spec.To
	}

	filters, err := toFilters(spec.Filters)
	if err != nil {
		return nil, err
	}

	limit := spec.Limit
	if limit <= 0 {
		limit = queryDefaultLimit
	}
	if limit > queryMaxLimit {
		limit = queryMaxLimit
	}

	switch spec.Chart {
	case domain.ChartMetric:
		n, err := s.store.Count(ctx, scope, filters)
		if err != nil {
			return nil, err
		}
		return &Result{Total: &n}, nil

	case domain.ChartCategory:
		buckets, err := s.store.TopValues(ctx, scope, spec.Dimension, filters, limit)
		if err != nil {
			return nil, err
		}
		return &Result{Buckets: toBuckets(buckets)}, nil

	case domain.ChartTime:
		interval, err := toInterval(spec.Interval)
		if err != nil {
			return nil, err
		}
		if spec.Dimension == "" {
			points, err := s.store.Timeline(ctx, scope, filters, interval)
			if err != nil {
				return nil, err
			}
			return &Result{Points: toPoints(points)}, nil
		}
		series, err := s.store.TimelineByField(ctx, scope, spec.Dimension, filters, interval, limit)
		if err != nil {
			return nil, err
		}
		return &Result{Series: toSeries(series)}, nil

	case domain.ChartTable:
		rows, total, err := s.store.FetchPage(ctx, scope, filters, store.Page{
			Limit:  limit,
			SortBy: "@timestamp",
			Order:  store.Desc,
		})
		if err != nil {
			return nil, err
		}
		return &Result{Rows: rows, Total: &total}, nil
	}

	return nil, domain.ErrUnknownChart
}

// The operators a widget filter may use. Every one of them is a store operator
// spelled the same way, so a spec reads like what the store will be asked.
var ops = map[string]store.Op{
	"eq": store.OpEq, "not_eq": store.OpNotEq,
	"in": store.OpIn, "not_in": store.OpNotIn,
	"gt": store.OpGt, "gte": store.OpGte,
	"lt": store.OpLt, "lte": store.OpLte,
	"between": store.OpBetween, "not_between": store.OpNotBetween,
	"contains": store.OpContains, "not_contains": store.OpNotContains,
	"starts_with": store.OpStartsWith, "not_starts_with": store.OpNotStartsWith,
	"ends_with": store.OpEndsWith, "not_ends_with": store.OpNotEndsWith,
	"exists": store.OpExists, "not_exists": store.OpNotExists,
}

func toFilters(in []domain.Filter) ([]store.Filter, error) {
	out := make([]store.Filter, 0, len(in))
	for _, f := range in {
		op, ok := ops[f.Op]
		if !ok {
			return nil, fmt.Errorf("%w: %s", domain.ErrUnknownOp, f.Op)
		}
		out = append(out, store.Filter{Field: f.Field, Op: op, Value: f.Value})
	}
	return out, nil
}

// toInterval maps the bucket sizes a dashboard offers. Calendar units are not
// the same as a fixed duration — a day is not always 24 hours — so the ones
// that have a calendar meaning use it.
func toInterval(s string) (store.Interval, error) {
	switch s {
	case "", "auto":
		return store.Interval{Duration: time.Hour}, nil
	case "1m":
		return store.Interval{Duration: time.Minute}, nil
	case "5m":
		return store.Interval{Duration: 5 * time.Minute}, nil
	case "15m":
		return store.Interval{Duration: 15 * time.Minute}, nil
	case "1h":
		return store.Interval{Calendar: store.CalendarHour}, nil
	case "1d":
		return store.Interval{Calendar: store.CalendarDay}, nil
	case "1w":
		return store.Interval{Calendar: store.CalendarWeek}, nil
	}
	return store.Interval{}, fmt.Errorf("unknown interval: %s", s)
}

var ErrNoTenantScope = errors.New("dashboards: no tenant in scope")

func scopeTenant(ctx context.Context) (string, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		if tenancy.Enabled() {
			return "", ErrNoTenantScope
		}
		return store.AllTenants, nil
	}
	return tenant, nil
}
