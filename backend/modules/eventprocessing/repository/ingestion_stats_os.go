package repository

import (
	"context"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

const statisticsIndexPattern = "v11-statistics-*"

type osIngestionStatsRepository struct{}

func NewIngestionStatsRepository() connectors.IngestionStatsRepository {
	return &osIngestionStatsRepository{}
}

// timeRangeFilter builds the bool filter shared by every stats query: a
// @timestamp range plus an optional doc `type` term.
func timeRangeFilter(statusType, from, to string) map[string]any {
	filters := []any{
		map[string]any{"range": map[string]any{
			"@timestamp": map[string]any{"gte": from, "lte": to},
		}},
	}
	if statusType != "" {
		filters = append(filters, map[string]any{"term": map[string]any{"type": statusType}})
	}
	return map[string]any{"bool": map[string]any{"filter": filters}}
}

var sumCount = map[string]any{"events": map[string]any{"sum": map[string]any{"field": "count"}}}

func (r *osIngestionStatsRepository) TotalsByField(ctx context.Context, field, statusType, from, to string, top int) ([]dto.IngestionStatsBucket, int64, error) {
	body := map[string]any{
		"size":  0,
		"query": timeRangeFilter(statusType, from, to),
		"aggs": map[string]any{
			"groups": map[string]any{
				"terms": map[string]any{"field": field + ".keyword", "size": top},
				"aggs": map[string]any{
					"events":    map[string]any{"sum": map[string]any{"field": "count"}},
					"last_seen": map[string]any{"max": map[string]any{"field": "@timestamp"}},
				},
			},
			"total": map[string]any{"sum": map[string]any{"field": "count"}},
		},
	}

	res, err := osdk.RawSearch(ctx, []string{statisticsIndexPattern}, body)
	if err != nil {
		return nil, 0, err
	}

	out := make([]dto.IngestionStatsBucket, 0)
	for _, b := range aggBuckets(res.Aggregations, "groups") {
		out = append(out, dto.IngestionStatsBucket{
			Key:      asString(b["key"]),
			Count:    subSum(b, "events"),
			LastSeen: subValueString(b, "last_seen"),
		})
	}
	return out, subSum(res.Aggregations, "total"), nil
}

func (r *osIngestionStatsRepository) Timeline(ctx context.Context, statusType, interval, from, to string) ([]dto.TimelinePoint, error) {
	body := map[string]any{
		"size":  0,
		"query": timeRangeFilter(statusType, from, to),
		"aggs": map[string]any{
			"timeline": dateHistogram(interval, from, to),
		},
	}

	res, err := osdk.RawSearch(ctx, []string{statisticsIndexPattern}, body)
	if err != nil {
		return nil, err
	}
	return timelinePoints(aggBuckets(res.Aggregations, "timeline")), nil
}

func dateHistogram(interval, from, to string) map[string]any {
	return map[string]any{
		"date_histogram": map[string]any{
			"field":           "@timestamp",
			"fixed_interval":  interval,
			"min_doc_count":   0,
			"extended_bounds": map[string]any{"min": from, "max": to},
		},
		"aggs": sumCount,
	}
}

func (r *osIngestionStatsRepository) TimelineByField(ctx context.Context, field, statusType, interval, from, to string, top int) ([]dto.TimelineSeries, error) {
	body := map[string]any{
		"size":  0,
		"query": timeRangeFilter(statusType, from, to),
		"aggs": map[string]any{
			"groups": map[string]any{
				"terms": map[string]any{"field": field + ".keyword", "size": top},
				"aggs": map[string]any{
					"timeline": dateHistogram(interval, from, to),
				},
			},
		},
	}

	res, err := osdk.RawSearch(ctx, []string{statisticsIndexPattern}, body)
	if err != nil {
		return nil, err
	}

	out := make([]dto.TimelineSeries, 0)
	for _, b := range aggBuckets(res.Aggregations, "groups") {
		out = append(out, dto.TimelineSeries{
			Key:    asString(b["key"]),
			Points: timelinePoints(subBuckets(b, "timeline")),
		})
	}
	return out, nil
}

// --- generic aggregation-map navigation (RawSearch returns map[string]any) ---

func aggBuckets(aggs map[string]any, name string) []map[string]any {
	agg, _ := aggs[name].(map[string]any)
	return toBucketSlice(agg)
}

func subBuckets(parent map[string]any, name string) []map[string]any {
	agg, _ := parent[name].(map[string]any)
	return toBucketSlice(agg)
}

func toBucketSlice(agg map[string]any) []map[string]any {
	raw, _ := agg["buckets"].([]any)
	out := make([]map[string]any, 0, len(raw))
	for _, b := range raw {
		if bm, ok := b.(map[string]any); ok {
			out = append(out, bm)
		}
	}
	return out
}

func timelinePoints(buckets []map[string]any) []dto.TimelinePoint {
	pts := make([]dto.TimelinePoint, 0, len(buckets))
	for _, b := range buckets {
		pts = append(pts, dto.TimelinePoint{
			Timestamp: asString(b["key_as_string"]),
			Count:     subSum(b, "events"),
		})
	}
	return pts
}

// subSum reads the {value: <number>} of a metric sub-aggregation as int64.
func subSum(parent map[string]any, name string) int64 {
	agg, _ := parent[name].(map[string]any)
	if v, ok := agg["value"].(float64); ok {
		return int64(v)
	}
	return 0
}

func subValueString(parent map[string]any, name string) string {
	agg, _ := parent[name].(map[string]any)
	return asString(agg["value_as_string"])
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
