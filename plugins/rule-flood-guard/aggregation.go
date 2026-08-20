package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"
)

const (
	falsePositiveTag = "False positive"
	statusOpen       = "Open"
	maxCombinations  = 10000
)

type ruleBucket struct {
	TenantID   string
	RuleName   string
	DataSource string
	Count      int64
}

// floodGroupFields is the grouping key. tenantId is not a grouping detail, it
// is the isolation: the driver renders these as a literal GROUP BY, so dropping
// it sums every tenant's alerts into one bucket and disables a rule on volume
// no single customer produced.
var floodGroupFields = []string{"tenantId", "name", "dataSource"}

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
		// One pass has to see the whole instance. Isolation comes from grouping
		// on tenantId below, not from narrowing this read — do not "fix" this
		// to a single tenant.
		Tenant:  store.AllTenants,
		Dataset: datasetAlerts,
		From:    now.Add(-window),
		To:      now,
	}

	groups, err := alertStore.GroupBy(ctx, scope,
		floodGroupFields,
		floodFilters(),
		store.GroupOpts{Limit: maxCombinations},
	)
	if err != nil {
		return nil, err
	}

	return bucketsFromGroups(groups), nil
}

func bucketsFromGroups(groups []store.Group) []ruleBucket {
	buckets := make([]ruleBucket, 0, len(groups))

	var walk func(level []store.Group, path map[string]string)
	walk = func(level []store.Group, path map[string]string) {
		for _, g := range level {
			next := make(map[string]string, len(path)+1)
			for k, v := range path {
				next[k] = v
			}
			next[g.Field] = g.Key

			if len(g.Children) == 0 {
				buckets = append(buckets, ruleBucket{
					TenantID:   next["tenantId"],
					RuleName:   next["name"],
					DataSource: next["dataSource"],
					Count:      g.Count,
				})
				continue
			}
			walk(g.Children, next)
		}
	}
	walk(groups, map[string]string{})

	return buckets
}
