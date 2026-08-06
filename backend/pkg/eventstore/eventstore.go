package eventstore

import (
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
		// Only logs keep the whole record as text, so only logs can be searched
		// that way. Asking it of alerts is an error rather than a query that
		// quietly matches nothing.
		TextColumns: map[store.Dataset]string{
			DatasetLogs: "raw",
		},
		TenantColumn: "tenantId",

		// What an index pattern used to encode besides the table:
		// v11-log-o365-* named the logs table and the o365 data type at once.
		DataTypeColumn: "dataType",
		TimeColumn:     "@timestamp",
	})
	if err != nil {
		return nil, err
	}

	return &Store{Driver: d, Conn: conn}, nil
}
