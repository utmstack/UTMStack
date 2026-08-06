package repository

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// Reader is the compliance module's slice of the event store — only Count is
// needed for the alert-activity dimension. Kept local so this module doesn't
// depend on the whole *eventstore.Store.
type Reader interface {
	Count(ctx context.Context, s store.Scope, filters []store.Filter) (int64, error)
}

type chAlerts struct{ store Reader }

// NewCHAlerts returns a counter over the alert dataset (activity dimension).
// A nil reader (event store not configured) is legal — CountByRuleNames returns
// 0 and lets the evaluator fall through to coverage-based status.
func NewCHAlerts(r Reader) connectors.OpenSearchAlerts { return &chAlerts{store: r} }

func (r *chAlerts) CountByRuleNames(ctx context.Context, ruleNames []string, sinceISO string) (int64, error) {
	if r.store == nil || len(ruleNames) == 0 {
		return 0, nil
	}
	scope := store.Scope{
		Tenant:  authz.TenantIDFromContext(ctx),
		Dataset: eventstore.DatasetAlerts,
	}
	if t, err := time.Parse(time.RFC3339, sinceISO); err == nil {
		scope.From = t
	}
	filters := []store.Filter{{Field: "name", Op: store.OpIn, Value: ruleNames}}
	return r.store.Count(ctx, scope, filters)
}
