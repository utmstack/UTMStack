package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
)

const enqueueSuccessType = "enqueue_success"

type chStatsReader struct{ conn driver.Conn }

func NewStatsReader(conn driver.Conn) connectors.StatsReader {
	if conn == nil {
		return nil
	}
	return &chStatsReader{conn: conn}
}

func (r *chStatsReader) DistinctSources(ctx context.Context, from, to time.Time) ([]connectors.StatSource, error) {
	const q = `
		SELECT tenantId, dataSource, dataType, max(` + "`@timestamp`" + `) AS lastSeen
		FROM statistics
		WHERE type = ?
		  AND ` + "`@timestamp`" + ` BETWEEN ? AND ?
		GROUP BY tenantId, dataSource, dataType`

	rows, err := r.conn.Query(ctx, q, enqueueSuccessType, from.UTC(), to.UTC())
	if err != nil {
		return nil, fmt.Errorf("datasources: reading ingestion statistics: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]connectors.StatSource, 0, 128)
	var unparseable int
	for rows.Next() {
		var (
			s      connectors.StatSource
			tenant string
		)
		if err := rows.Scan(&tenant, &s.DataSource, &s.DataType, &s.LastSeen); err != nil {
			return nil, err
		}
		if s.DataSource == "" {
			continue
		}

		id, err := uuid.Parse(tenant)
		if err != nil {
			unparseable++
			continue
		}
		s.TenantID = id
		out = append(out, s)
	}
	if unparseable > 0 {
		_ = catcher.Error("datasources: ingestion statistics carried unusable tenants", nil,
			map[string]any{"rows": unparseable})
	}
	return out, rows.Err()
}

// TenantUsageByDay sums the same "enqueue_success" counters DistinctSources
// reads, grouped by tenant instead of by data source, for one UTC calendar
// day. day's time-of-day is ignored — only its UTC date matters.
func (r *chStatsReader) TenantUsageByDay(ctx context.Context, day time.Time) ([]connectors.TenantUsage, error) {
	const q = `
		SELECT tenantId, sum(count) AS events, sum(bytes) AS bytes
		FROM statistics
		WHERE type = ?
		  AND toDate(` + "`@timestamp`" + `) = toDate(?)
		GROUP BY tenantId`

	rows, err := r.conn.Query(ctx, q, enqueueSuccessType, day.UTC())
	if err != nil {
		return nil, fmt.Errorf("datasources: reading tenant usage: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]connectors.TenantUsage, 0, 16)
	var unparseable int
	for rows.Next() {
		var (
			u      connectors.TenantUsage
			tenant string
		)
		if err := rows.Scan(&tenant, &u.EventCount, &u.Bytes); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(tenant)
		if err != nil {
			unparseable++
			continue
		}
		u.TenantID = id
		out = append(out, u)
	}
	if unparseable > 0 {
		_ = catcher.Error("datasources: tenant usage carried unusable tenants", nil,
			map[string]any{"rows": unparseable})
	}
	return out, rows.Err()
}
