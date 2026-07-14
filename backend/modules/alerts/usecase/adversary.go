package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// adversaryIndexPattern is the OpenSearch index the adversary aggregation reads.
const adversaryIndexPattern = "v11-alert-*"

// adversaryBucketAggs is the set of sub-aggregations run under each adversary
// bucket (by host or by IP) — shared so both grouping strategies fetch the
// same shape of data.
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
					},
				},
			},
		},
	}
}

type adversaryUsecase struct{}

func NewAdversaryUsecase() connectors.AdversaryUsecase {
	return &adversaryUsecase{}
}

func (u *adversaryUsecase) FetchAdversaryAlerts(
	ctx context.Context,
	filters []common_models.FilterType,
) ([]dto.AdversaryResponse, error) {
	exists, err := osdk.IndexExists(ctx, adversaryIndexPattern)
	if err != nil {
		_ = catcher.Error("alerts adversary: IndexExists failed", err, nil)
		return nil, err
	}
	if !exists {
		return []dto.AdversaryResponse{}, nil
	}

	body := map[string]any{
		"size":  0,
		"query": common_models.FiltersToQuery(filters),
		"aggs": map[string]any{
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
		},
	}

	res, err := osdk.RawSearch(ctx, []string{adversaryIndexPattern}, body)
	if err != nil {
		_ = catcher.Error("alerts adversary: OpenSearch search failed", err, nil)
		return nil, err
	}

	// parseAdversaryAggs decodes the full {"aggregations": {...}} envelope.
	rawAggs, err := json.Marshal(map[string]any{"aggregations": res.Aggregations})
	if err != nil {
		return nil, err
	}
	return parseAdversaryAggs(rawAggs)
}

// ---------------------------------------------------------------------------
// Aggregation response structs
// ---------------------------------------------------------------------------

type adversarySearchResponse struct {
	Aggregations struct {
		ByHost adversaryFilterAgg `json:"by_host"`
		ByIP   adversaryFilterAgg `json:"by_ip"`
	} `json:"aggregations"`
}

// adversaryFilterAgg wraps the filter agg's own "grouped" terms sub-agg.
type adversaryFilterAgg struct {
	Grouped struct {
		Buckets []adversaryBucket `json:"buckets"`
	} `json:"grouped"`
}

type adversaryBucket struct {
	Key          string `json:"key"`
	AdversaryObj struct {
		Hits struct {
			Hits []adversaryHitSource `json:"hits"`
		} `json:"hits"`
	} `json:"adversary_obj"`
	Alerts struct {
		AlertsHits struct {
			Hits struct {
				Hits []adversaryHitSource `json:"hits"`
			} `json:"hits"`
		} `json:"alerts_hits"`
	} `json:"alerts"`
	ChildAlerts struct {
		Buckets []adversaryChildBucket `json:"buckets"`
	} `json:"child_alerts"`
}

type adversaryChildBucket struct {
	Key       string `json:"key"`
	ChildHits struct {
		Hits struct {
			Hits []adversaryHitSource `json:"hits"`
		} `json:"hits"`
	} `json:"child_hits"`
}

type adversaryHitSource struct {
	Source json.RawMessage `json:"_source"`
}

// ---------------------------------------------------------------------------

func parseAdversaryAggs(rawResp []byte) ([]dto.AdversaryResponse, error) {
	var sr adversarySearchResponse
	if err := json.Unmarshal(rawResp, &sr); err != nil {
		return nil, fmt.Errorf("alerts adversary: decode aggregation response: %w", err)
	}

	buckets := make([]adversaryBucket, 0, len(sr.Aggregations.ByHost.Grouped.Buckets)+len(sr.Aggregations.ByIP.Grouped.Buckets))
	buckets = append(buckets, sr.Aggregations.ByHost.Grouped.Buckets...)
	buckets = append(buckets, sr.Aggregations.ByIP.Grouped.Buckets...)

	groups := make([]dto.AdversaryResponse, 0, len(buckets))

	for _, bucket := range buckets {
		// Extract the adversary Side from the top_hits result.
		if len(bucket.AdversaryObj.Hits.Hits) == 0 || bucket.AdversaryObj.Hits.Hits[0].Source == nil {
			continue
		}
		var wrapper dto.AdversaryWrapper
		if err := json.Unmarshal(bucket.AdversaryObj.Hits.Hits[0].Source, &wrapper); err != nil {
			_ = catcher.Error("alerts adversary: decode adversary wrapper", err, nil)
			continue
		}

		// Collect parent alerts.
		parentHits := bucket.Alerts.AlertsHits.Hits.Hits
		parents := make([]domain.UtmAlert, 0, len(parentHits))
		for _, h := range parentHits {
			if h.Source == nil {
				continue
			}
			var alert domain.UtmAlert
			if err := json.Unmarshal(h.Source, &alert); err != nil {
				_ = catcher.Error("alerts adversary: decode parent alert", err, nil)
				continue
			}
			if alert.ParentID == "" {
				parents = append(parents, alert)
			}
		}

		childMap := make(map[string][]domain.UtmAlert)
		for _, cb := range bucket.ChildAlerts.Buckets {
			children := make([]domain.UtmAlert, 0, len(cb.ChildHits.Hits.Hits))
			for _, h := range cb.ChildHits.Hits.Hits {
				if h.Source == nil {
					continue
				}
				var child domain.UtmAlert
				if err := json.Unmarshal(h.Source, &child); err != nil {
					_ = catcher.Error("alerts adversary: decode child alert", err, nil)
					continue
				}
				children = append(children, child)
			}
			childMap[cb.Key] = children
		}

		alertsWithChildren := make([]dto.AlertWithChildren, 0, len(parents))
		for _, parent := range parents {
			kids := childMap[parent.ID]
			if kids == nil {
				kids = []domain.UtmAlert{}
			}
			alertsWithChildren = append(alertsWithChildren, dto.AlertWithChildren{
				Alert:    parent,
				Children: kids,
			})
		}

		groups = append(groups, dto.AdversaryResponse{
			Adversary: wrapper.Adversary,
			Alerts:    alertsWithChildren,
		})
	}

	return groups, nil
}
