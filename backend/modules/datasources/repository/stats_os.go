package repository

import (
	"context"
	"time"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
)

const (
	statisticsIndexPattern = "v11-statistics-*"
	// enqueueSuccessType is the doc `type` for logs that successfully ingested —
	// the only ones that should register a source (dropped logs don't count).
	enqueueSuccessType = "enqueue_success"
	maxSourceBuckets   = 10000
	maxTypeBuckets     = 100
)

type osStatsReader struct{}

func NewStatsReader() connectors.StatsReader { return &osStatsReader{} }

// DistinctSources aggregates v11-statistics-* over [from,to] into the set of
// (dataSource, dataType) pairs that ingested, each with its last @timestamp.
func (r *osStatsReader) DistinctSources(ctx context.Context, from, to time.Time) ([]connectors.StatSource, error) {
	body := map[string]any{
		"size": 0,
		"query": map[string]any{"bool": map[string]any{"filter": []any{
			map[string]any{"range": map[string]any{"@timestamp": map[string]any{
				"gte": from.UTC().Format(time.RFC3339),
				"lte": to.UTC().Format(time.RFC3339),
			}}},
			map[string]any{"term": map[string]any{"type": enqueueSuccessType}},
		}}},
		"aggs": map[string]any{
			"sources": map[string]any{
				"terms": map[string]any{"field": "dataSource.keyword", "size": maxSourceBuckets},
				"aggs": map[string]any{
					"types": map[string]any{
						"terms": map[string]any{"field": "dataType.keyword", "size": maxTypeBuckets},
						"aggs": map[string]any{
							"last_seen": map[string]any{"max": map[string]any{"field": "@timestamp"}},
						},
					},
				},
			},
		},
	}

	res, err := osdk.RawSearch(ctx, []string{statisticsIndexPattern}, body)
	if err != nil {
		return nil, err
	}

	out := make([]connectors.StatSource, 0)
	for _, src := range aggBuckets(res.Aggregations, "sources") {
		dataSource := asString(src["key"])
		if dataSource == "" {
			continue
		}
		for _, t := range subBuckets(src, "types") {
			dataType := asString(t["key"])
			if dataType == "" {
				continue
			}
			out = append(out, connectors.StatSource{
				DataSource: dataSource,
				DataType:   dataType,
				LastSeen:   subMaxTime(t, "last_seen"),
			})
		}
	}
	return out, nil
}

// --- aggregation-map navigation (RawSearch returns map[string]any) ---

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

// subMaxTime reads a max-over-@timestamp sub-aggregation, preferring the
// formatted value_as_string and falling back to the epoch-millis value.
func subMaxTime(parent map[string]any, name string) time.Time {
	agg, _ := parent[name].(map[string]any)
	if s, ok := agg["value_as_string"].(string); ok && s != "" {
		if ts, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return ts.UTC()
		}
		if ts, err := time.Parse(time.RFC3339, s); err == nil {
			return ts.UTC()
		}
	}
	if v, ok := agg["value"].(float64); ok && v > 0 {
		return time.UnixMilli(int64(v)).UTC()
	}
	return time.Time{}
}

func asString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
