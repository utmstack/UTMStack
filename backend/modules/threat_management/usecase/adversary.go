package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/threat_management/dto"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	ospkg "github.com/utmstack/utmstack/backend/pkg/opensearch"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

const alertIndexPattern = "v11-alert-*"

type AdversaryUsecase struct {
	osClient *ospkg.Client
}

func NewAdversaryUsecase(osClient *ospkg.Client) *AdversaryUsecase {
	return &AdversaryUsecase{osClient: osClient}
}

func (u *AdversaryUsecase) FetchAdversaryAlerts(
	ctx context.Context,
	filters []common_models.FilterType,
) ([]dto.AdversaryResponse, error) {
	exists, err := u.osClient.IndexExists(ctx, alertIndexPattern)
	if err != nil {
		logger.Error("threat_management: IndexExists failed: " + err.Error())
		return nil, err
	}
	if !exists {
		return []dto.AdversaryResponse{}, nil
	}

	body := map[string]any{
		"size":  0,
		"query": common_models.FiltersToQuery(filters),
		"aggs": map[string]any{
			"adversary": map[string]any{
				"terms": map[string]any{
					"field": "adversary.host.keyword",
					"size":  100,
				},
				"aggs": map[string]any{
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
				},
			},
		},
	}

	rawResp, status, err := u.osClient.RawSearch(ctx, alertIndexPattern, body)
	if err != nil {
		logger.Error("threat_management: OpenSearch search failed: " + err.Error())
		return nil, err
	}
	if status < 200 || status >= 300 {
		logger.Error(fmt.Sprintf("threat_management: unexpected status %d from OpenSearch", status))
		return nil, fmt.Errorf("opensearch returned status %d", status)
	}

	return parseAdversaryAggs(rawResp)
}

// ---------------------------------------------------------------------------
// Aggregation response structs
// ---------------------------------------------------------------------------

type searchResponse struct {
	Aggregations struct {
		Adversary struct {
			Buckets []adversaryBucket `json:"buckets"`
		} `json:"adversary"`
	} `json:"aggregations"`
}

type adversaryBucket struct {
	Key          string `json:"key"`
	AdversaryObj struct {
		Hits struct {
			Hits []hitSource `json:"hits"`
		} `json:"hits"`
	} `json:"adversary_obj"`
	Alerts struct {
		AlertsHits struct {
			Hits struct {
				Hits []hitSource `json:"hits"`
			} `json:"hits"`
		} `json:"alerts_hits"`
	} `json:"alerts"`
	ChildAlerts struct {
		Buckets []childBucket `json:"buckets"`
	} `json:"child_alerts"`
}

type childBucket struct {
	Key       string `json:"key"`
	ChildHits struct {
		Hits struct {
			Hits []hitSource `json:"hits"`
		} `json:"hits"`
	} `json:"child_hits"`
}

type hitSource struct {
	Source json.RawMessage `json:"_source"`
}

// ---------------------------------------------------------------------------

func parseAdversaryAggs(rawResp []byte) ([]dto.AdversaryResponse, error) {
	var sr searchResponse
	if err := json.Unmarshal(rawResp, &sr); err != nil {
		return nil, fmt.Errorf("threat_management: decode aggregation response: %w", err)
	}

	groups := make([]dto.AdversaryResponse, 0, len(sr.Aggregations.Adversary.Buckets))

	for _, bucket := range sr.Aggregations.Adversary.Buckets {
		// Extract the adversary Side from the top_hits result.
		if len(bucket.AdversaryObj.Hits.Hits) == 0 || bucket.AdversaryObj.Hits.Hits[0].Source == nil {
			continue
		}
		var wrapper dto.AdversaryWrapper
		if err := json.Unmarshal(bucket.AdversaryObj.Hits.Hits[0].Source, &wrapper); err != nil {
			logger.Error("threat_management: decode adversary wrapper: " + err.Error())
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
				logger.Error("threat_management: decode parent alert: " + err.Error())
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
					logger.Error("threat_management: decode child alert: " + err.Error())
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
