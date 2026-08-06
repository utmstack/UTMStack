package usecase

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

// adversaryBucketAggs is the set of sub-aggregations run under each adversary
// bucket (by host or by IP) — shared so both grouping strategies fetch the
// same shape of data.

type adversaryUsecase struct {
	repo connectors.AdversaryRepository
}

func NewAdversaryUsecase(repo connectors.AdversaryRepository) connectors.AdversaryUsecase {
	return &adversaryUsecase{repo: repo}
}

func (u *adversaryUsecase) FetchAdversaryAlerts(
	ctx context.Context,
	filters []common_models.FilterType,
) ([]dto.AdversaryResponse, error) {
	rawAggs, err := u.repo.AdversaryAggs(ctx, filters)
	if err != nil {
		return nil, err
	}
	if rawAggs == nil {
		return []dto.AdversaryResponse{}, nil
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
