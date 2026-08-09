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
