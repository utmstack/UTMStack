package repository

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

type Reader interface {
	Count(ctx context.Context, s store.Scope, filters []store.Filter) (int64, error)
	TopValues(ctx context.Context, s store.Scope, field string, filters []store.Filter, n int) ([]store.Bucket, error)
}

type chEvents struct{ store Reader }

func NewEventCounter(r Reader) connectors.EventCounter { return &chEvents{store: r} }

const ruleNameField = "name"

func (r *chEvents) scope(ctx context.Context, dataset domain.Dataset, dataType string, from, to time.Time) (store.Scope, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		return store.Scope{}, store.ErrNoTenant
	}
	ds := eventstore.DatasetLogs
	if dataset == domain.DatasetAlerts {
		ds = eventstore.DatasetAlerts
	}
	return store.Scope{
		Tenant:   tenant,
		Dataset:  ds,
		DataType: dataType,
		From:     from,
		To:       to,
	}, nil
}

func (r *chEvents) Count(ctx context.Context, q connectors.CheckQuery) (int64, error) {
	if r.store == nil {
		return 0, nil
	}
	scope, err := r.scope(ctx, q.Dataset, q.DataType, q.From, q.To)
	if err != nil {
		return 0, err
	}
	filters, err := common_models.ToStoreFilters(q.Filters)
	if err != nil {
		return 0, err
	}
	return r.store.Count(ctx, scope, filters)
}

func (r *chEvents) HasData(ctx context.Context, dataset domain.Dataset, dataType string, from, to time.Time) (bool, error) {
	if r.store == nil {
		return false, nil
	}
	scope, err := r.scope(ctx, dataset, dataType, from, to)
	if err != nil {
		return false, err
	}
	n, err := r.store.Count(ctx, scope, nil)
	return n > 0, err
}

func (r *chEvents) CountByRuleNames(ctx context.Context, ruleNames []string, from, to time.Time) (map[string]int64, error) {
	out := make(map[string]int64, len(ruleNames))
	if r.store == nil || len(ruleNames) == 0 {
		return out, nil
	}
	scope, err := r.scope(ctx, domain.DatasetAlerts, "", from, to)
	if err != nil {
		return nil, err
	}
	filters := []store.Filter{{Field: ruleNameField, Op: store.OpIn, Value: ruleNames}}
	buckets, err := r.store.TopValues(ctx, scope, ruleNameField, filters, len(ruleNames))
	if err != nil {
		return nil, err
	}
	for _, b := range buckets {
		out[b.Key] = b.Count
	}
	return out, nil
}
