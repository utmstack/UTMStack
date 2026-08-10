package main

import (
	"context"
	"encoding/json"
	"time"

	sdkos "github.com/threatwinds/go-sdk/os"
)

const (
	alertIndex           = "v11-alert-*"
	falsePositiveTag     = "False positive"
	ruleNameAgg          = "by_rule_name"
	dataSourceAgg        = "by_data_source"
	maxDataSourceBuckets = 10000
)

type ruleBucket struct {
	RuleName   string
	DataSource string
	Count      int64
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
				"aggs": map[string]any{
					dataSourceAgg: map[string]any{
						"terms": map[string]any{
							"field": "dataSource.keyword",
							"size":  maxDataSourceBuckets,
						},
					},
				},
			},
		},
	}
}

type termsBucketsResponse struct {
	Buckets []struct {
		Key      string `json:"key"`
		DocCount int64  `json:"doc_count"`
		Sub      struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"by_data_source"`
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
		for _, sub := range bucket.Sub.Buckets {
			buckets = append(buckets, ruleBucket{RuleName: bucket.Key, DataSource: sub.Key, Count: sub.DocCount})
		}
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
