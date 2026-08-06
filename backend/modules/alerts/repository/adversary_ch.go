package repository

import (
	"context"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

const (
	adversaryGroups = 100
	adversarySample = 100
)

type chAdversaryRepo struct {
	store *eventstore.Store
}

func NewAdversaryRepository(s *eventstore.Store) connectors.AdversaryRepository {
	return &chAdversaryRepo{store: s}
}

func (r *chAdversaryRepo) AdversaryGroups(ctx context.Context, filters []common_models.FilterType) ([]connectors.AdversaryGroup, error) {
	scope, err := alertScope(ctx)
	if err != nil {
		return nil, err
	}

	scope, filters = common_models.SplitTimeBounds(scope, filters)

	base, err := common_models.ToStoreFilters(filters)
	if err != nil {
		return nil, err
	}
	opts := store.GroupOpts{Limit: adversaryGroups, TopHits: adversarySample}

	byHost, err := r.store.GroupBy(ctx, scope, []string{"adversary.host"},
		append(base, store.Filter{Field: "adversary.host", Op: store.OpExists}), opts)
	if err != nil {
		return nil, err
	}

	byIP, err := r.store.GroupBy(ctx, scope, []string{"adversary.ip"},
		append(base,
			store.Filter{Field: "adversary.host", Op: store.OpNotExists},
			store.Filter{Field: "adversary.ip", Op: store.OpExists},
		), opts)
	if err != nil {
		return nil, err
	}

	out := make([]connectors.AdversaryGroup, 0, len(byHost)+len(byIP))
	for _, g := range append(byHost, byIP...) {
		if g.Key == "" {
			continue
		}
		out = append(out, connectors.AdversaryGroup{Key: g.Key, Alerts: g.Hits})
	}
	return out, nil
}
