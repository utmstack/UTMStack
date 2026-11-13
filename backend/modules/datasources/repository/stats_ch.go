package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

const enqueueSuccessType = "enqueue_success"

// The statistics table moved to the event store with the rest of the pipeline,
// so these reads go through its Query helper: FetchQuery takes no bound
// parameters, and the only values inlined are constants and UTC-formatted
// times. Rows come back as JSON, and timestamps are rendered with
// formatDateTime — a bare DateTime column does not round-trip through JSON.
type chStatsReader struct {
	store *eventstore.Store
}

func NewStatsReader(s *eventstore.Store) connectors.StatsReader {
	if s == nil {
		return nil
	}
	return &chStatsReader{store: s}
}

func (r *chStatsReader) DistinctSources(ctx context.Context, from, to time.Time) ([]connectors.StatSource, error) {
	q := fmt.Sprintf(`
		SELECT tenantId, dataSource, dataType, formatDateTime(max(`+"`@timestamp`"+`), '%%Y-%%m-%%dT%%H:%%M:%%SZ') AS lastSeen
		FROM %s
		WHERE type = '%s'
		  AND `+"`@timestamp`"+` BETWEEN '%s' AND '%s'
		GROUP BY tenantId, dataSource, dataType`,
		r.store.TableName(eventstore.DatasetStats),
		enqueueSuccessType,
		from.UTC().Format(time.RFC3339),
		to.UTC().Format(time.RFC3339))

	rows, err := r.store.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("datasources: reading ingestion statistics: %w", err)
	}

	out := make([]connectors.StatSource, 0, 128)
	var unparseable int
	for _, row := range rows {
		var s struct {
			Tenant   string `json:"tenantId"`
			Source   string `json:"dataSource"`
			Type     string `json:"dataType"`
			LastSeen string `json:"lastSeen"`
		}
		if err := json.Unmarshal(row, &s); err != nil {
			return nil, err
		}
		if s.Source == "" {
			continue
		}
		id, err := uuid.Parse(s.Tenant)
		if err != nil {
			unparseable++
			continue
		}
		last, err := time.Parse(time.RFC3339, s.LastSeen)
		if err != nil {
			continue
		}
		out = append(out, connectors.StatSource{
			TenantID:   id,
			DataSource: s.Source,
			DataType:   s.Type,
			LastSeen:   last.UTC(),
		})
	}
	if unparseable > 0 {
		_ = catcher.Error("datasources: ingestion statistics carried unusable tenants", nil,
			map[string]any{"rows": unparseable})
	}
	return out, nil
}

// TenantUsageByDay sums the same "enqueue_success" counters DistinctSources
// reads, grouped by tenant instead of by data source, for one UTC calendar
// day. day's time-of-day is ignored — only its UTC date matters.
func (r *chStatsReader) TenantUsageByDay(ctx context.Context, day time.Time) ([]connectors.TenantUsage, error) {
	q := fmt.Sprintf(`
		SELECT tenantId, sum(count) AS events, sum(bytes) AS bytes
		FROM %s
		WHERE type = '%s'
		  AND toDate(`+"`@timestamp`"+`) = toDate('%s')
		GROUP BY tenantId`,
		r.store.TableName(eventstore.DatasetStats),
		enqueueSuccessType,
		day.UTC().Format(time.RFC3339))

	rows, err := r.store.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("datasources: reading tenant usage: %w", err)
	}

	out := make([]connectors.TenantUsage, 0, 16)
	var unparseable int
	for _, row := range rows {
		var u struct {
			Tenant string  `json:"tenantId"`
			Events *uint64 `json:"events"`
			Bytes  *uint64 `json:"bytes"`
		}
		if err := json.Unmarshal(row, &u); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(u.Tenant)
		if err != nil {
			unparseable++
			continue
		}
		ten := connectors.TenantUsage{TenantID: id}
		if u.Events != nil {
			ten.EventCount = int64(*u.Events)
		}
		if u.Bytes != nil {
			ten.Bytes = int64(*u.Bytes)
		}
		out = append(out, ten)
	}
	if unparseable > 0 {
		_ = catcher.Error("datasources: tenant usage carried unusable tenants", nil,
			map[string]any{"rows": unparseable})
	}
	return out, nil
}
