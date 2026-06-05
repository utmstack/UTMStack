package usecase

import (
	"context"
	"encoding/json"
	"time"

	osdk "github.com/threatwinds/go-sdk/os"
)

const statisticsIndexPattern = "v11-statistics-*"

// SourceActivity is the projection of a successful enqueue event for a given
// (dataSource, dataType). Mirrors the Java StatisticDocument fields the AssetSync
// service actually consumed.
type SourceActivity struct {
	DataSource string
	DataType   string
	Timestamp  time.Time
}

// SourceActivityProvider queries OpenSearch for the latest enqueue_success events
// since a checkpoint. Mirrors the Java SourceActivityProvider behaviour.
type SourceActivityProvider struct{}

func NewSourceActivityProvider() *SourceActivityProvider { return &SourceActivityProvider{} }

// Fetch returns one entry per (dataSource, dataType) reflecting the most recent activity
// strictly after `since`. The returned slice can be safely empty.
func (p *SourceActivityProvider) Fetch(ctx context.Context, since time.Time) ([]SourceActivity, error) {
	body := map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"filter": []any{
					map[string]any{"term": map[string]any{"type": "enqueue_success"}},
					map[string]any{"range": map[string]any{"@timestamp": map[string]any{"gt": since.UTC().Format(time.RFC3339)}}},
				},
			},
		},
		"aggs": map[string]any{
			"by_source": map[string]any{
				"terms": map[string]any{"field": "dataSource.keyword", "size": 1000},
				"aggs": map[string]any{
					"by_type": map[string]any{
						"terms": map[string]any{"field": "dataType.keyword", "size": 100},
						"aggs": map[string]any{
							"latest": map[string]any{
								"top_hits": map[string]any{
									"size": 1,
									"sort": []any{
										map[string]any{"@timestamp": map[string]any{"order": "desc"}},
									},
									"_source": []string{"dataSource", "dataType", "@timestamp"},
								},
							},
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
	return extractActivities(res.Aggregations)
}

func extractActivities(aggs map[string]any) ([]SourceActivity, error) {
	out := []SourceActivity{}
	bySource, _ := aggs["by_source"].(map[string]any)
	if bySource == nil {
		return out, nil
	}
	sourceBuckets, _ := bySource["buckets"].([]any)
	for _, sb := range sourceBuckets {
		sm, _ := sb.(map[string]any)
		if sm == nil {
			continue
		}
		byType, _ := sm["by_type"].(map[string]any)
		if byType == nil {
			continue
		}
		typeBuckets, _ := byType["buckets"].([]any)
		for _, tb := range typeBuckets {
			tm, _ := tb.(map[string]any)
			if tm == nil {
				continue
			}
			latest, _ := tm["latest"].(map[string]any)
			if latest == nil {
				continue
			}
			hitsOuter, _ := latest["hits"].(map[string]any)
			if hitsOuter == nil {
				continue
			}
			hits, _ := hitsOuter["hits"].([]any)
			if len(hits) == 0 {
				continue
			}
			h0, _ := hits[0].(map[string]any)
			if h0 == nil {
				continue
			}
			source, _ := h0["_source"].(map[string]any)
			if source == nil {
				continue
			}
			ds, _ := source["dataSource"].(string)
			dt, _ := source["dataType"].(string)
			tsStr, _ := source["@timestamp"].(string)
			if ds == "" || dt == "" {
				continue
			}
			ts := parseAnyTimestamp(tsStr)
			out = append(out, SourceActivity{
				DataSource: ds,
				DataType:   dt,
				Timestamp:  ts,
			})
		}
	}
	return out, nil
}

func parseAnyTimestamp(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	// Try millisecond epoch number first.
	var n json.Number
	if err := json.Unmarshal([]byte(s), &n); err == nil {
		if i, err := n.Int64(); err == nil && i > 1e10 {
			return time.UnixMilli(i).UTC()
		}
	}
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
