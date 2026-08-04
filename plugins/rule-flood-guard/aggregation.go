package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"
)

const (
	falsePositiveTag = "False positive"
	statusOpen       = 2
	maxCombinations  = 10000
)

type ruleBucket struct {
	RuleName   string
	DataSource string
	Count      int64
}

func floodFilters() []store.Filter {
	return []store.Filter{
		{Field: "status", Op: store.OpEq, Value: statusOpen},
		{Field: "tags", Op: store.OpNotContains, Value: falsePositiveTag},
		{Field: "parentId", Op: store.OpEq, Value: ""},
	}
}

func searchRuleBuckets(ctx context.Context, window time.Duration) ([]ruleBucket, error) {
	now := time.Now().UTC()
	scope := store.Scope{
		Tenant:  store.AllTenants,
		Dataset: datasetAlerts,
		From:    now.Add(-window),
		To:      now,
	}

	groups, err := alertStore.GroupBy(ctx, scope,
		[]string{"name", "dataSource"},
		floodFilters(),
		store.GroupOpts{Limit: maxCombinations},
	)
	if err != nil {
		return nil, err
	}
	buckets := make([]ruleBucket, 0, len(groups))
	for _, byName := range groups {
		for _, bySource := range byName.Children {
			buckets = append(buckets, ruleBucket{
				RuleName:   byName.Key,
				DataSource: bySource.Key,
				Count:      bySource.Count,
			})
		}
	}

	return buckets, nil
}
