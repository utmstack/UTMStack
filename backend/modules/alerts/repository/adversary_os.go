package repository

import (
	"context"
	"encoding/json"

	"github.com/threatwinds/go-sdk/catcher"
	osdk "github.com/threatwinds/go-sdk/os"

	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// adversaryIndex is where the adversary roll-up reads from.
const adversaryIndex = "v11-alert-*"

// alertSourceIncludes trims top_hits _source to only the fields the frontend
// renders, so the roll-up does not ship whole alerts it will not use.
var alertSourceIncludes = []string{"id", "name", "target", "severity"}

type osAdversaryRepo struct{}

func NewAdversaryRepository() *osAdversaryRepo { return &osAdversaryRepo{} }

// AdversaryAggs groups alerts by attacker and returns the raw aggregation
// envelope. A nil result means there is nothing to read yet.
//
// It lives here rather than in the usecase so it goes through osAggregate and
// is scoped to the acting tenant like every other read in this module. Built in
// the usecase it reached OpenSearch directly, and the roll-up counted every
// tenant's alerts for whoever asked.
func (r *osAdversaryRepo) AdversaryAggs(ctx context.Context, filters []common_models.FilterType) (json.RawMessage, error) {
	exists, err := osdk.IndexExists(ctx, adversaryIndex)
	if err != nil {
		_ = catcher.Error("alerts adversary: IndexExists failed", err, nil)
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	aggs := map[string]any{
		// Group by adversary.host when present — internally-identified
		// attackers (compromised hosts, agents) carry a hostname.
		"by_host": map[string]any{
			"filter": map[string]any{
				"exists": map[string]any{"field": "adversary.host"},
			},
			"aggs": map[string]any{
				"grouped": map[string]any{
					"terms": map[string]any{
						"field": "adversary.host.keyword",
						"size":  100,
					},
					"aggs": adversaryBucketAggs(),
				},
			},
		},
		// Fall back to adversary.ip for docs with no host — the common
		// case for Internet-facing threats (malicious IPs, external
		// scanners), which are identified by IP, not hostname.
		"by_ip": map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": []map[string]any{
						{"exists": map[string]any{"field": "adversary.host"}},
					},
					"must": []map[string]any{
						{"exists": map[string]any{"field": "adversary.ip"}},
					},
				},
			},
			"aggs": map[string]any{
				"grouped": map[string]any{
					"terms": map[string]any{
						"field": "adversary.ip.keyword",
						"size":  100,
					},
					"aggs": adversaryBucketAggs(),
				},
			},
		},
	}

	raw, err := osAggregate(ctx, adversaryIndex, common_models.FiltersToQuery(filters), aggs, "30s")
	if err != nil {
		_ = catcher.Error("alerts adversary: OpenSearch search failed", err, nil)
		return nil, err
	}
	return raw, nil
}

func adversaryBucketAggs() map[string]any {
	return map[string]any{
		// Grab the most-recent adversary Side object.
		"adversary_obj": map[string]any{
			"top_hits": map[string]any{
				"size": 1,
				"sort": []map[string]any{
					{"@timestamp": map[string]any{"order": "desc"}},
				},
				"_source": map[string]any{
					"includes": []string{"adversary"},
				},
			},
		},
		// Parent alerts: must_not exists parentId.
		"alerts": map[string]any{
			"filter": map[string]any{
				"bool": map[string]any{
					"must_not": []map[string]any{
						{"exists": map[string]any{"field": "parentId"}},
					},
				},
			},
			"aggs": map[string]any{
				"alerts_hits": map[string]any{
					"top_hits": map[string]any{
						"size": 100,
						"sort": []map[string]any{
							{"@timestamp": map[string]any{"order": "desc"}},
						},
						"_source": map[string]any{
							"includes": alertSourceIncludes,
						},
					},
				},
			},
		},
		// Child alerts grouped by parentId.
		"child_alerts": map[string]any{
			"terms": map[string]any{
				"field": "parentId.keyword",
				"size":  50,
			},
			"aggs": map[string]any{
				"child_hits": map[string]any{
					"top_hits": map[string]any{
						"size": 50,
						"sort": []map[string]any{
							{"@timestamp": map[string]any{"order": "desc"}},
						},
						"_source": map[string]any{
							"includes": alertSourceIncludes,
						},
					},
				},
			},
		},
	}
}
