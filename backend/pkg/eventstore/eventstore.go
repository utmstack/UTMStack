package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/clickhouse"
	"github.com/threatwinds/go-sdk/store/factory"
	"github.com/utmstack/utmstack/backend/pkg/env"
)

const (
	DatasetLogs        store.Dataset = "logs"
	DatasetAlerts      store.Dataset = "alerts"
	DatasetStats       store.Dataset = "statistics"
	DatasetUserHistory store.Dataset = "user_profile_history"
	MigrationPah       string        = "file://migrations/clickhouse"
)

type Store struct {
	store.Store
	built *factory.Built
}

func New() (*Store, error) {

	config_path := env.String("STORAGE_CONFIG_PATH", "", false)
	if config_path == "" {
		return nil, nil
	}

	raw, err := os.ReadFile(config_path)
	if err != nil {
		return nil, err
	}
	cfg, err := factory.Parse(raw)

	if err != nil {
		return nil, err
	}

	retry_intervals := []uint32{5,10,15}

	var store *Store
	for _, wait_time := range retry_intervals{

		built, err := factory.New(cfg, factory.Options{
			OnSecondaryError: func(op string, err error) {
				catcher.Error(fmt.Sprintf("secondary storage %s start error: ",op),err,nil )
			},
		})

		ctx := context.Background()
		store = &Store{Store: built.Store, built: built}

		err = store.initTables(ctx)
		if err == nil{
			break
		}
		catcher.Log(fmt.Sprintf("clickhouse not ready retrying on %d s, reason: ",wait_time), map[string]any{"error":err})
		time.Sleep(time.Second*time.Duration(wait_time))
	}


	if err != nil {
		return nil, err
	}

	catcher.Info("✅ clickhouse ready — migrations complete", nil)
	return store, nil

}

func (s *Store) GetReader() store.Reader {
	return s
}

func (s *Store) GetWritter() store.Reader {
	return s
}

func (s *Store) GetAdmin() store.Reader {
	return s
}

func (s *Store) Close() {
	s.built.Close()
}



func (s *Store) initTables(ctx context.Context) error {
	db := env.String("CLICKHOUSE_DB", "utmstack", false)


	args := map[string]string{
		"db":db,
		"logs":string(DatasetLogs),
		"alerts":string(DatasetAlerts),
		"stats":string(DatasetStats),
		"user_profile":string(DatasetUserHistory),
	}

	d, ok := s.Store.(*clickhouse.Driver)
	if !ok {
		catcher.Log("Clickhouse not configured", map[string]any{})
		return nil // not CH, nothing to create
	}
	h, err := d.Health(ctx)
	if err != nil {
		return err
	}
	if h.Status != store.HealthOK {
		catcher.Log("CH unhealthy , skipping table init", map[string]any{"status": h.Status})
		return fmt.Errorf("CH unhealthy")
	}



	migrations, err := loadMigrations("migrations/clickhouse",args)
	if err != nil {
		return err
	}

	catcher.Info("running clickhouse migrations",nil)
	for _, q := range migrations {
		if err := d.Exec(ctx, q.Content); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) TableName(d store.Dataset) string {
	name := map[store.Dataset]string{DatasetLogs: "logs", DatasetAlerts: "alerts", DatasetStats: "statistics", DatasetUserHistory: "user_profile_history"}[d]
	db := env.String("CLICKHOUSE_DB", "utmstack", false)
	return db + "." + name
}

// makes sql queries into the reader storage
func (s *Store) Query(ctx context.Context, sql string) ([]json.RawMessage, error) {
	return s.GetReader().FetchQuery(ctx, store.Scope{Tenant: store.AllTenants}, sql)
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
		if _, err := s.Drop(ctx, store.Scope{Tenant: tenantID, Dataset: d}); err != nil {
			return fmt.Errorf("purge %s: %w", d, err)
		}
	}
	return nil
}

func SupportsTextSearch(d store.Dataset) bool { return d == DatasetLogs }
