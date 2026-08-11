package repository

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ErrNoTenantScope = errors.New("eventprocessing: no tenant in scope")

type chIngestionStatsRepository struct {
	store *eventstore.Store
}

func NewIngestionStatsRepository(s *eventstore.Store) connectors.IngestionStatsRepository {
	return &chIngestionStatsRepository{store: s}
}

func statsScope(ctx context.Context) (store.Scope, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		if tenancy.Enabled() {
			return store.Scope{}, ErrNoTenantScope
		}
		tenant = store.AllTenants
	}
	return store.Scope{Tenant: tenant, Dataset: eventstore.DatasetStats}, nil
}

func (r *chIngestionStatsRepository) prepare(ctx context.Context, filters []common_models.FilterType) (store.Scope, []store.Filter, error) {
	scope, err := statsScope(ctx)
	if err != nil {
		return store.Scope{}, nil, err
	}
	scope, filters = common_models.SplitTimeBounds(scope, filters)

	base, err := common_models.ToStoreFilters(filters)
	if err != nil {
		return store.Scope{}, nil, err
	}
	return scope, base, nil
}

func (r *chIngestionStatsRepository) TotalsByField(ctx context.Context, field string, filters []common_models.FilterType, top int) ([]dto.IngestionStatsBucket, int64, error) {
	scope, base, err := r.prepare(ctx, filters)
	if err != nil {
		return nil, 0, err
	}

	groups, err := r.store.GroupBy(ctx, scope, []string{field}, base,
		store.GroupOpts{Limit: top, TopHits: 1, SortBy: "@timestamp", Order: store.Desc})
	if err != nil {
		return nil, 0, err
	}

	buckets := make([]dto.IngestionStatsBucket, 0, len(groups))
	for _, g := range groups {
		if g.Key == "" {
			continue
		}
		buckets = append(buckets, dto.IngestionStatsBucket{
			Key:      g.Key,
			Count:    g.Count,
			LastSeen: lastSeenOf(g.Hits),
		})
	}

	total, err := r.store.Count(ctx, scope, base)
	if err != nil {
		return nil, 0, err
	}
	return buckets, total, nil
}

func (r *chIngestionStatsRepository) Timeline(ctx context.Context, filters []common_models.FilterType, interval string) ([]dto.TimelinePoint, error) {
	scope, base, err := r.prepare(ctx, filters)
	if err != nil {
		return nil, err
	}

	points, err := r.store.Timeline(ctx, scope, base, parseInterval(interval))
	if err != nil {
		return nil, err
	}
	return toPoints(points), nil
}

func (r *chIngestionStatsRepository) TimelineByField(ctx context.Context, field string, filters []common_models.FilterType, interval string, top int) ([]dto.TimelineSeries, error) {
	scope, base, err := r.prepare(ctx, filters)
	if err != nil {
		return nil, err
	}

	series, err := r.store.TimelineByField(ctx, scope, field, base, parseInterval(interval), top)
	if err != nil {
		return nil, err
	}

	out := make([]dto.TimelineSeries, 0, len(series))
	for _, s := range series {
		out = append(out, dto.TimelineSeries{Key: s.Key, Points: toPoints(s.Points)})
	}
	return out, nil
}

func toPoints(in []store.Point) []dto.TimelinePoint {
	out := make([]dto.TimelinePoint, 0, len(in))
	for _, p := range in {
		out = append(out, dto.TimelinePoint{
			Timestamp: p.At.UTC().Format(time.RFC3339),
			Count:     p.Count,
		})
	}
	return out
}

func lastSeenOf(hits []json.RawMessage) string {
	if len(hits) == 0 {
		return ""
	}
	var doc struct {
		Timestamp string `json:"@timestamp"`
	}
	if err := json.Unmarshal(hits[0], &doc); err != nil {
		return ""
	}
	return doc.Timestamp
}

func parseInterval(s string) store.Interval {
	s = strings.TrimSpace(s)
	if s == "1M" || strings.EqualFold(s, "month") {
		return store.Interval{Calendar: store.CalendarMonth}
	}
	switch strings.ToLower(s) {
	case "1d", "day":
		return store.Interval{Calendar: store.CalendarDay}
	case "1w", "week":
		return store.Interval{Calendar: store.CalendarWeek}
	}
	if d, err := time.ParseDuration(s); err == nil && d > 0 {
		return store.Interval{Duration: d}
	}
	return store.Interval{Duration: time.Hour}
}
