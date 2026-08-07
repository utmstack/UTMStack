package repository

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ruleOperators = map[string]store.Op{
	"filter_term":    store.OpEq,
	"must_not_term":  store.OpNotEq,
	"filter_match":   store.OpContains,
	"must_not_match": store.OpNotContains,
}

func logScope(ctx context.Context) (store.Scope, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		if tenancy.Enabled() {
			return store.Scope{}, ErrNoTenantScope
		}
		tenant = store.AllTenants
	}
	return store.Scope{Tenant: tenant, Dataset: eventstore.DatasetLogs}, nil
}

func (r *chAlertRepo) RelatedLogRefs(ctx context.Context, steps []domain.CorrelationStep, anchorTS time.Time, maxSize int) ([]domain.LogRef, bool, error) {
	seen := make(map[string]struct{})
	var out []domain.LogRef
	truncated := false

	for i := range steps {
		refs, total, _, err := r.runStep(ctx, &steps[i], anchorTS, maxSize)
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

func (r *chAlertRepo) runStep(ctx context.Context, step *domain.CorrelationStep, anchorTS time.Time, maxSize int) ([]domain.LogRef, int64, bool, error) {
	filters := make([]store.Filter, 0, len(step.With))
	for _, e := range step.With {
		if e.Value == nil {
			continue
		}
		op, ok := ruleOperators[e.Operator]
		if !ok {
			continue
		}
		filters = append(filters, store.Filter{Field: e.Field, Op: op, Value: e.Value})
	}

	scope, err := logScope(ctx)
	if err != nil {
		return nil, 0, false, err
	}
	scope.To = anchorTS
	if step.Within != "" {
		if d, err := time.ParseDuration(step.Within); err == nil && d > 0 {
			scope.From = anchorTS.Add(-d)
		}
	}

	total, err := r.store.Count(ctx, scope, filters)
	if err != nil {
		return nil, 0, false, err
	}

	if uint64(total) >= step.Count {
		buckets, err := r.store.TopValues(ctx, scope, "id", filters, maxSize)
		if err != nil {
			return nil, 0, false, err
		}
		refs := make([]domain.LogRef, 0, len(buckets))
		for _, b := range buckets {
			if b.Key == "" {
				continue
			}
			refs = append(refs, domain.LogRef{ID: b.Key})
		}
		return refs, total, true, nil
	}

	var refs []domain.LogRef
	var fired bool
	for i := range step.Or {
		orRefs, _, orFired, err := r.runStep(ctx, &step.Or[i], anchorTS, maxSize)
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
