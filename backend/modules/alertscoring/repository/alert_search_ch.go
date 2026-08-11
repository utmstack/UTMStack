package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/threatwinds/go-sdk/store"

	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

var ErrNoTenantScope = errors.New("alertscoring: no tenant in scope")

type chAlertSearch struct {
	store *eventstore.Store
}

func NewAlertSearch(s *eventstore.Store) connectors.AlertSearch {
	return &chAlertSearch{store: s}
}

func scope(ctx context.Context) (store.Scope, error) {
	tenant := authz.TenantIDFromContext(ctx)
	if tenant == "" {
		if tenancy.Enabled() {
			return store.Scope{}, ErrNoTenantScope
		}
		tenant = store.AllTenants
	}
	return store.Scope{Tenant: tenant, Dataset: eventstore.DatasetAlerts}, nil
}

func (r *chAlertSearch) prepare(ctx context.Context, filters []common_models.FilterType) (store.Scope, []store.Filter, error) {
	s, err := scope(ctx)
	if err != nil {
		return store.Scope{}, nil, err
	}
	s, filters = common_models.SplitTimeBounds(s, filters)

	out, err := common_models.ToStoreFilters(filters)
	if err != nil {
		return store.Scope{}, nil, err
	}
	return s, out, nil
}

func (r *chAlertSearch) FetchByID(ctx context.Context, id string) (*alertdomain.UtmAlert, error) {
	s, err := scope(ctx)
	if err != nil {
		return nil, err
	}

	raw, err := r.store.FetchByID(ctx, s, id)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}

	var a alertdomain.UtmAlert
	if err := json.Unmarshal(raw, &a); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *chAlertSearch) Count(ctx context.Context, filters []common_models.FilterType) (int64, error) {
	s, f, err := r.prepare(ctx, filters)
	if err != nil {
		return 0, err
	}
	return r.store.Count(ctx, s, f)
}

func (r *chAlertSearch) Recent(ctx context.Context, filters []common_models.FilterType, n int, oldestFirst bool) ([]alertdomain.UtmAlert, error) {
	s, f, err := r.prepare(ctx, filters)
	if err != nil {
		return nil, err
	}

	page := store.Page{Limit: n, SortBy: "@timestamp", Order: store.Desc}
	if oldestFirst {
		page.Order = store.Asc
	}

	rows, _, err := r.store.FetchPage(ctx, s, f, page)
	if err != nil {
		return nil, err
	}

	out := make([]alertdomain.UtmAlert, 0, len(rows))
	for _, raw := range rows {
		var a alertdomain.UtmAlert
		if err := json.Unmarshal(raw, &a); err != nil {
			continue
		}
		out = append(out, a)
	}
	return out, nil
}
