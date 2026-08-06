package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
)

// enqueueSuccessType is the statistics row a log that actually ingested
// produces. Dropped logs write their own type and must not register a source.
const enqueueSuccessType = "enqueue_success"

type chStatsReader struct{ conn driver.Conn }

func NewStatsReader(conn driver.Conn) connectors.StatsReader {
	if conn == nil {
		return nil
	}
	return &chStatsReader{conn: conn}
}

// DistinctSources reports what ingested in the window, per tenant, with the
// last time each was seen.
//
// This is written by hand rather than through the store driver because the
// driver has no shape for it: its GroupBy returns counts, and what liveness
// needs is the newest timestamp per group. The window is a day, so taking "now"
// for anything that appeared in it would report a source that stopped this
// morning as live.
//
// It reads across tenants on purpose — one reconciler serves every tenant —
// which is why the tenant is selected rather than filtered on, and why each row
// carries its own downstream.
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
	for rows.Next() {
		var s connectors.StatSource
		if err := rows.Scan(&s.TenantID, &s.DataSource, &s.DataType, &s.LastSeen); err != nil {
			return nil, err
		}
		if s.DataSource == "" || s.TenantID == "" {
			continue
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
