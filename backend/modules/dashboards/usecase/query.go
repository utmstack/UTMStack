package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
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

type Result struct {
	Total   *int64            `json:"total,omitempty"`
	Buckets []store.Bucket    `json:"buckets,omitempty"`
	Points  []store.Point     `json:"points,omitempty"`
	Series  []store.Series    `json:"series,omitempty"`
	Rows    []json.RawMessage `json:"rows,omitempty"`
}

func (s *QueryService) Run(ctx context.Context, spec domain.Spec) (*Result, error) {
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	scope := store.Scope{
		Tenant:   authz.TenantIDFromContext(ctx),
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
		return &Result{Buckets: buckets}, nil

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
			return &Result{Points: points}, nil
		}
		series, err := s.store.TimelineByField(ctx, scope, spec.Dimension, filters, interval, limit)
		if err != nil {
			return nil, err
		}
		return &Result{Series: series}, nil

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

var ops = map[string]store.Op{
	"eq": store.OpEq, "not_eq": store.OpNotEq,
	"in": store.OpIn, "not_in": store.OpNotIn,
	"gt": store.OpGt, "gte": store.OpGte,
	"lt": store.OpLt, "lte": store.OpLte,
	"between":  store.OpBetween,
	"contains": store.OpContains, "not_contains": store.OpNotContains,
	"exists": store.OpExists,
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
