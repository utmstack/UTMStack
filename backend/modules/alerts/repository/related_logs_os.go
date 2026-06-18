package repository

import (
	"context"
	"encoding/json"
	"time"

	osdk "github.com/threatwinds/go-sdk/os"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
)

const relatedProcess = "alerts_related_logs"

func (r *osAlertRepo) GetRawByID(ctx context.Context, alertID string) (json.RawMessage, error) {
	body := map[string]any{
		"size":  1,
		"query": map[string]any{"ids": map[string]any{"values": []string{alertID}}},
	}
	res, err := osdk.RawSearch(ctx, []string{alertIndex}, body)
	if err != nil {
		return nil, err
	}
	if len(res.Hits.Hits) == 0 {
		return nil, nil
	}
	return json.Marshal(res.Hits.Hits[0].Source)
}

func (r *osAlertRepo) RelatedLogRefs(ctx context.Context, steps []domain.CorrelationStep, anchorTS time.Time, maxSize int) ([]domain.LogRef, bool, error) {
	seen := make(map[string]struct{})
	var out []domain.LogRef
	truncated := false

	for i := range steps {
		refs, total, _, err := r.runRelatedStep(ctx, &steps[i], anchorTS, maxSize)
		if err != nil {
			return nil, false, err
		}
		if total > int64(len(refs)) && len(refs) >= maxSize {
			truncated = true
		}
		for _, ref := range refs {
			if _, dup := seen[ref.ID]; dup {
				continue
			}
			seen[ref.ID] = struct{}{}
			out = append(out, ref)
		}
	}
	return out, truncated, nil
}

func (r *osAlertRepo) runRelatedStep(ctx context.Context, step *domain.CorrelationStep, anchorTS time.Time, maxSize int) ([]domain.LogRef, int64, bool, error) {
	idx := step.IndexPattern
	if idx == "" {
		return nil, 0, false, nil
	}

	builder := osdk.NewQueryBuilder(ctx, []string{idx}, relatedProcess)

	if step.Within != "" {
		if d, err := time.ParseDuration(step.Within); err == nil {
			builder.Range("@timestamp", "gte", anchorTS.Add(-d).Format(time.RFC3339))
			builder.Range("@timestamp", "lte", anchorTS.Format(time.RFC3339))
		}
	}

	for _, e := range step.With {
		if e.Value == nil {
			continue
		}
		switch e.Operator {
		case "filter_term":
			builder.Term(e.Field, e.Value)
		case "must_not_term":
			builder.MustNot(*osdk.NewQueryBuilder(ctx, []string{idx}, relatedProcess).Term(e.Field, e.Value).Build().Query)
		case "filter_match":
			if s, ok := e.Value.(string); ok {
				builder.Match(e.Field, s)
			}
		case "must_not_match":
			if s, ok := e.Value.(string); ok {
				builder.MustNot(*osdk.NewQueryBuilder(ctx, []string{idx}, relatedProcess).Match(e.Field, s).Build().Query)
			}
		}
	}

	sr := builder.Size(int64(maxSize)).Build()
	res, err := sr.WideSearchIn(ctx, []string{idx})
	if err != nil {
		return nil, 0, false, err
	}

	total := res.Hits.Total.Value
	if uint64(total) >= step.Count {
		return hitsToRefs(res.Hits.Hits), total, true, nil
	}

	var refs []domain.LogRef
	var fired bool
	for i := range step.Or {
		orRefs, _, orFired, err := r.runRelatedStep(ctx, &step.Or[i], anchorTS, maxSize)
		if err != nil {
			return nil, 0, false, err
		}
		if orFired {
			refs = append(refs, orRefs...)
			fired = true
		}
	}
	return refs, total, fired && len(step.Or) != 0, nil
}

func hitsToRefs(hits []osdk.Hit) []domain.LogRef {
	refs := make([]domain.LogRef, 0, len(hits))
	for _, h := range hits {
		refs = append(refs, domain.LogRef{ID: h.ID, Index: h.Index})
	}
	return refs
}
