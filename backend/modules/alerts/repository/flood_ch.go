package repository

import (
	"context"
	"time"

	"github.com/utmstack/go-sdk/v2/store"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

const (
	floodMaxCombinations = 10000
	floodFalsePositive   = "False positive"
	floodStatusOpen      = "Open"
)

// floodGroupFields is the grouping key. tenantId is not a detail: the driver renders these as a
// literal GROUP BY, so leaving it out would add every tenant's alerts into one bucket and disable a
// rule over a volume no single customer produced.
var floodGroupFields = []string{"tenantId", "name", "dataSource"}

type chFloodRepo struct {
	store *eventstore.Store
}

func NewFloodRepository(s *eventstore.Store) connectors.FloodRepository {
	return &chFloodRepo{store: s}
}

func floodFilters() []store.Filter {
	return []store.Filter{
		{Field: "status", Op: store.OpEq, Value: floodStatusOpen},
		{Field: "tags", Op: store.OpNotContains, Value: floodFalsePositive},
		{Field: "parentId", Op: store.OpEq, Value: ""},
	}
}

func (r *chFloodRepo) FloodBuckets(ctx context.Context, window time.Duration) ([]connectors.FloodBucket, error) {
	now := time.Now().UTC()

	// One pass has to see the whole instance. Isolation comes from grouping on tenantId, not from
	// narrowing this read to a tenant.
	scope := store.Scope{
		Tenant:  store.AllTenants,
		Dataset: eventstore.DatasetAlerts,
		From:    now.Add(-window),
		To:      now,
	}

	groups, err := r.store.GroupBy(ctx, scope, floodGroupFields, floodFilters(), store.GroupOpts{Limit: floodMaxCombinations})
	if err != nil {
		return nil, err
	}

	return floodBucketsFromGroups(groups), nil
}

// floodBucketsFromGroups flattens the nested groups the store returns, one level per grouping field,
// into one bucket per combination.
func floodBucketsFromGroups(groups []store.Group) []connectors.FloodBucket {
	buckets := make([]connectors.FloodBucket, 0, len(groups))

	var walk func(level []store.Group, path map[string]string)

	walk = func(level []store.Group, path map[string]string) {
		for _, g := range level {
			next := make(map[string]string, len(path)+1)
			for k, v := range path {
				next[k] = v
			}

			next[g.Field] = g.Key

			if len(g.Children) == 0 {
				buckets = append(buckets, connectors.FloodBucket{
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
