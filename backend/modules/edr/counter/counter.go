package counter

import (
	"context"
	"errors"
	"time"

	"github.com/threatwinds/go-sdk/store"
	"github.com/utmstack/utmstack/backend/modules/edr/connectors"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

// Detection count over a window in the logs store. dataType is the
// EDR spool's constant; see agent/edr/event/event.go.
const edrDataType = "utmstack_edr"

type logsCounter struct {
	store *eventstore.Store
}

func NewLogsCounter(s *eventstore.Store) connectors.DetectionCounter {
	return &logsCounter{store: s}
}

// CountInWindow counts EDR events between from and to. A store without
// ClickHouse (nil) is a valid install shape — it reports zero detections,
// not an error, so the overview still answers.
func (c *logsCounter) CountInWindow(ctx context.Context, from, to time.Time) (int64, error) {
	if c.store == nil || c.store.Driver == nil {
		return 0, nil
	}
	scope := store.Scope{
		Tenant:   authz.TenantIDFromContext(ctx),
		Dataset:  eventstore.DatasetLogs,
		DataType: edrDataType,
		From:     from,
		To:       to,
	}
	n, err := c.store.Count(ctx, scope, nil)
	if err != nil {
		if errors.Is(err, store.ErrUnsupported) {
			return 0, nil
		}
		return 0, err
	}
	return n, nil
}
