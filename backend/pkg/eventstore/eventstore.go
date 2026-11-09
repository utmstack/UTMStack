package eventstore

import (
	"context"
	"fmt"
	"time"

	chdriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/clickhouse"

	"github.com/utmstack/utmstack/backend/pkg/env"
)

const (
	DatasetLogs        = store.Dataset("logs")
	DatasetAlerts      = store.Dataset("alerts")
	DatasetStats       = store.Dataset("statistics")
	DatasetUserHistory = store.Dataset("user_profile_history")
)

type Store struct {
	*clickhouse.Driver
	history *clickhouse.Driver
	Conn    driver.Conn
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
		// Statistics is absent on purpose: it has no id column, and it is
		// counted rather than paged through, so it never pays the cost this
		// avoids.
		IDColumns: map[store.Dataset]string{
			DatasetLogs:   "id",
			DatasetAlerts: "id",
		},
		TenantColumn:   "tenantId",
		DataTypeColumn: "dataType",
		TimeColumn:     "@timestamp",
	})
	if err != nil {
		return nil, err
	}

	// The history table is keyed by snapshotted_at, not @timestamp, so it needs
	// its own driver for the retention math to pick the right column.
	h, err := clickhouse.New(clickhouse.Config{
		Conn:         conn,
		Database:     env.String("CLICKHOUSE_DB", "utmstack", false),
		Tables:       map[store.Dataset]string{DatasetUserHistory: "user_profile_history"},
		TenantColumn: "tenantId",
		TimeColumn:   "snapshotted_at",
	})
	if err != nil {
		return nil, err
	}

	return &Store{Driver: d, history: h, Conn: conn}, nil
}

// DriverFor returns the clickhouse driver that owns the given dataset. The
// history dataset has its own driver because its time column differs from the
// log tables; everything else shares the main one.
func (s *Store) DriverFor(ds store.Dataset) *clickhouse.Driver {
	if ds == DatasetUserHistory {
		return s.history
	}
	return s.Driver
}

func (s *Store) TableName(d store.Dataset) string {
	name := map[store.Dataset]string{DatasetLogs: "logs", DatasetAlerts: "alerts", DatasetStats: "statistics", DatasetUserHistory: "user_profile_history"}[d]
	db := env.String("CLICKHOUSE_DB", "utmstack", false)
	return db + "." + name
}

// PurgeTenant deletes all rows scoped to the given tenant across every
// dataset table. ClickHouse mutation, not synchronous — the ALTER returns
// once accepted and the rows disappear as the mutation runs.
//
// tenantID must be a UUID string (caller passes uuid.UUID.String()); it is
// inlined into the SQL because ClickHouse does not bind parameters inside
// ALTER ... DELETE mutation expressions. The UUID shape makes inlining safe.
func (s *Store) PurgeTenant(ctx context.Context, tenantID string) error {
	if _, err := uuid.Parse(tenantID); err != nil {
		return fmt.Errorf("purge: invalid tenant id: %w", err)
	}
	for _, d := range []store.Dataset{DatasetLogs, DatasetAlerts, DatasetStats, DatasetUserHistory} {
		sql := fmt.Sprintf("ALTER TABLE %s DELETE WHERE tenantId = '%s'", s.TableName(d), tenantID)
		if err := s.Conn.Exec(ctx, sql); err != nil {
			return fmt.Errorf("purge %s: %w", d, err)
		}
	}
	return nil
}

func SupportsTextSearch(d store.Dataset) bool { return d == DatasetLogs }
