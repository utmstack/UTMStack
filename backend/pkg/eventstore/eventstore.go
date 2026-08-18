package eventstore

import (
	"context"
	"fmt"
	"time"

	chdriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/clickhouse"

	"github.com/utmstack/utmstack/backend/pkg/env"
)

const (
	DatasetLogs   = store.Dataset("logs")
	DatasetAlerts = store.Dataset("alerts")
	DatasetStats  = store.Dataset("statistics")
)

type Store struct {
	*clickhouse.Driver
	Conn driver.Conn
}

func New() (*Store, error) {
	host := env.String("CLICKHOUSE_HOST", "", false)
	if host == "" {
		return nil, nil
	}

	conn, err := chdriver.Open(&chdriver.Options{
		Addr: []string{host + ":" + env.String("CLICKHOUSE_PORT", "9000", false)},
		Auth: chdriver.Auth{
			Database: env.String("CLICKHOUSE_DB", "utmstack", false),
			Username: env.String("CLICKHOUSE_USER", "default", false),
			Password: env.String("CLICKHOUSE_PASSWORD", "", false),
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	d, err := clickhouse.New(clickhouse.Config{
		Conn:     conn,
		Database: env.String("CLICKHOUSE_DB", "utmstack", false),
		Tables: map[store.Dataset]string{
			DatasetLogs:   "logs",
			DatasetAlerts: "alerts",
			DatasetStats:  "statistics",
		},
		TextColumns: map[store.Dataset]string{
			DatasetLogs: "raw",
		},
		TenantColumn:   "tenantId",
		DataTypeColumn: "dataType",
		TimeColumn:     "@timestamp",
	})
	if err != nil {
		return nil, err
	}

	return &Store{Driver: d, Conn: conn}, nil
}

func (s *Store) TableName(d store.Dataset) string {
	name := map[store.Dataset]string{DatasetLogs: "logs", DatasetAlerts: "alerts", DatasetStats: "statistics"}[d]
	db := env.String("CLICKHOUSE_DB", "utmstack", false)
	return db + "." + name
}

// PurgeTenant deletes all rows scoped to the given tenant across every
// dataset table. ClickHouse mutation, not synchronous — the ALTER returns
// once accepted and the rows disappear as the mutation runs.
func (s *Store) PurgeTenant(ctx context.Context, tenantID string) error {
	for _, d := range []store.Dataset{DatasetLogs, DatasetAlerts, DatasetStats} {
		sql := fmt.Sprintf("ALTER TABLE %s DELETE WHERE tenantId = ?", s.TableName(d))
		if err := s.Conn.Exec(ctx, sql, tenantID); err != nil {
			return fmt.Errorf("purge %s: %w", d, err)
		}
	}
	return nil
}

func SupportsTextSearch(d store.Dataset) bool { return d == DatasetLogs }
