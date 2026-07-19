package main

import (
	"context"
	"encoding/json"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
)

const alertIndex = "v11-alert-*"

const falsePositiveTag = "False positive"

const ruleNameAgg = "by_rule_name"

type ruleBucket struct {
	RuleName string
	Count    int64
}

func buildFloodQuery(window time.Duration, now time.Time) map[string]any {
	gte := now.Add(-window).UTC().Format(time.RFC3339)
	return map[string]any{
		"size": 0,
		"query": map[string]any{
			"bool": map[string]any{
				"must": []map[string]any{
					{"term": map[string]any{"status": 2}},
					{"range": map[string]any{"@timestamp": map[string]any{"gte": gte}}},
				},
				"must_not": []map[string]any{
					{"term": map[string]any{"tags": falsePositiveTag}},
					{"exists": map[string]any{"field": "parentId.keyword"}},
				},
			},
		},
		"aggs": map[string]any{
			ruleNameAgg: map[string]any{
				"terms": map[string]any{
					"field": "name.keyword",
					"size":  1000,
				},
			},
		},
	}
}

type termsBucketsResponse struct {
	Buckets []struct {
		Key      string `json:"key"`
		DocCount int64  `json:"doc_count"`
	} `json:"buckets"`
}

func parseRuleBuckets(aggs map[string]interface{}) ([]ruleBucket, error) {
	raw, ok := aggs[ruleNameAgg]
	if !ok {
		return nil, nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var parsed termsBucketsResponse
	if err := json.Unmarshal(b, &parsed); err != nil {
		return nil, err
	}
	buckets := make([]ruleBucket, 0, len(parsed.Buckets))
	for _, bucket := range parsed.Buckets {
		buckets = append(buckets, ruleBucket{RuleName: bucket.Key, Count: bucket.DocCount})
	}
	return buckets, nil
}

func searchRuleBuckets(ctx context.Context, window time.Duration) ([]ruleBucket, error) {
	query := buildFloodQuery(window, time.Now())
	result, err := sdkos.RawSearch(ctx, []string{alertIndex}, query)
	if err != nil {
		return nil, err
	}
	return parseRuleBuckets(result.Aggregations)
}
