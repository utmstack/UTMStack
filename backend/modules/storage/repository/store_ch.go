package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/threatwinds/go-sdk/store"
	"github.com/threatwinds/go-sdk/store/clickhouse"

	"github.com/utmstack/utmstack/backend/modules/storage/connectors"
	"github.com/utmstack/utmstack/backend/modules/storage/domain"
	"github.com/utmstack/utmstack/backend/pkg/eventstore"
)

type chStoreRepository struct {
	store *eventstore.Store
}

func NewStoreRepository(s *eventstore.Store) connectors.StoreRepository {
	return &chStoreRepository{store: s}
}

var datasets = map[domain.Dataset]store.Dataset{
	domain.DatasetLogs:        eventstore.DatasetLogs,
	domain.DatasetAlerts:      eventstore.DatasetAlerts,
	domain.DatasetStats:       eventstore.DatasetStats,
	domain.DatasetUserProfile: eventstore.DatasetUserHistory,
}

func (r *chStoreRepository) Retention(ctx context.Context, d domain.Dataset) (domain.Retention, error) {
	ds, ok := datasets[d]
	if !ok {
		return domain.Retention{}, domain.ErrUnknownDataset
	}

	got, err := r.store.Retention(ctx, ds)
	if err != nil {
		return domain.Retention{}, err
	}

	out := domain.Retention{
		Dataset:  d,
		KeepDays: days(got.Keep),
		ColdDays: days(got.ColdAfter),
	}
	// A table with no TTL keeps everything; saying so as "0 days" would read as
	// "delete immediately".
	if out.KeepDays == 0 {
		out.KeepDays = domain.MaxKeepDays
	}
	return out, nil
}

func (r *chStoreRepository) SetRetention(ctx context.Context, want domain.Retention) error {
	ds, ok := datasets[want.Dataset]
	if !ok {
		return domain.ErrUnknownDataset
	}

	// Rewriting a tiered TTL as a flat one would drop the move clause, and
	// nothing would report it: records would simply stop leaving the local disk.
	if !want.Tiered() {
		current, err := r.Retention(ctx, want.Dataset)
		if err != nil {
			return err
		}
		if current.Tiered() {
			return domain.ErrTieringPermanent
		}
	}

	return r.store.SetRetention(ctx, ds, toStore(want))
}

func (r *chStoreRepository) AdoptTiering(ctx context.Context, want domain.Retention) error {
	ds, ok := datasets[want.Dataset]
	if !ok {
		return domain.ErrUnknownDataset
	}
	return r.chDriver().EnableTiering(ctx, ds, domain.PolicyName, toStore(want))
}

func (r *chStoreRepository) Usage(ctx context.Context) ([]domain.Usage, error) {
	rows, err := r.store.Usage(ctx)
	if err != nil {
		return nil, err
	}
	hRows, err := r.store.Usage(ctx)
	if err != nil {
		return nil, err
	}
	rows = append(rows, hRows...)

	byDataset := make(map[domain.Dataset]*domain.Usage, len(rows))
	for _, u := range rows {
		d := domain.Dataset(u.Dataset)
		if !d.Valid() {
			continue
		}
		acc, ok := byDataset[d]
		if !ok {
			acc = &domain.Usage{Dataset: d}
			byDataset[d] = acc
		}
		acc.Documents += u.Documents
		acc.Bytes += u.Bytes
	}

	out := make([]domain.Usage, 0, len(datasets))
	for _, d := range domain.Datasets() {
		u := domain.Usage{Dataset: d}
		if acc, ok := byDataset[d]; ok {
			u = *acc
		}
		oldest, newest, err := r.span(ctx, d)
		if err != nil {
			return nil, err
		}
		u.Oldest, u.Newest = oldest, newest
		out = append(out, u)
	}
	return out, nil
}

// span reads how far back a dataset goes from the part metadata rather than the
// records, so it costs nothing on a table with years in it.
func (r *chStoreRepository) span(ctx context.Context, d domain.Dataset) (time.Time, time.Time, error) {
	rows, err := r.store.Query(ctx,
		"SELECT formatDateTime(min(min_time), '%Y-%m-%dT%H:%M:%SZ') AS oldest, formatDateTime(max(max_time), '%Y-%m-%dT%H:%M:%SZ') AS newest FROM system.parts WHERE database = currentDatabase() AND table = '"+string(d)+"' AND active")
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if len(rows) == 0 {
		return time.Time{}, time.Time{}, nil
	}
	var span struct {
		Oldest *time.Time `json:"oldest"`
		Newest *time.Time `json:"newest"`
	}
	if err := json.Unmarshal(rows[0], &span); err != nil {
		return time.Time{}, time.Time{}, err
	}
	if span.Oldest == nil || span.Newest == nil {
		return time.Time{}, time.Time{}, nil
	}
	return span.Oldest.UTC(), span.Newest.UTC(), nil
}

func (r *chStoreRepository) Health(ctx context.Context) (domain.Health, error) {
	h, err := r.store.Health(ctx)
	if err != nil {
		return domain.Health{}, err
	}
	return domain.Health{
		Status:      string(h.Status),
		DiskUsedPct: h.DiskUsedPct,
		Message:     h.Message,
	}, nil
}

func (r *chStoreRepository) ColdStorageReady(ctx context.Context) (bool, error) {
	rows, err := r.store.Query(ctx,
		"SELECT count() AS n FROM system.storage_policies WHERE policy_name = '"+domain.PolicyName+"' AND volume_name = '"+domain.ColdVolume+"'")
	if err != nil {
		return false, err
	}
	if len(rows) == 0 {
		return false, nil
	}
	var row struct {
		N uint64 `json:"n"`
	}
	if err := json.Unmarshal(rows[0], &row); err != nil {
		return false, err
	}
	return row.N > 0, nil
}

func (r *chStoreRepository) ColdDisk(ctx context.Context) (domain.ColdDiskInfo, error) {
	rows, err := r.store.Query(ctx,
		"SELECT count() AS n, max(free_space) AS free, max(is_read_only) AS ro FROM system.disks WHERE name = '"+domain.ColdDisk+"'")
	if err != nil {
		return domain.ColdDiskInfo{}, err
	}
	if len(rows) == 0 {
		return domain.ColdDiskInfo{}, nil
	}
	var row struct {
		N    uint64  `json:"n"`
		Free *uint64 `json:"free"`
		Ro   *uint8  `json:"ro"`
	}
	if err := json.Unmarshal(rows[0], &row); err != nil {
		return domain.ColdDiskInfo{}, err
	}
	free := uint64(0)
	if row.Free != nil {
		free = *row.Free
	}
	ro := uint8(0)
	if row.Ro != nil {
		ro = *row.Ro
	}
	return domain.ColdDiskInfo{Present: row.N > 0, ReadOnly: ro > 0, FreeBytes: free}, nil
}

func (r *chStoreRepository) ReloadConfig(ctx context.Context) error {
	d := r.chDriver()
	if d == nil {
		return nil
	}
	return d.Exec(ctx, "SYSTEM RELOAD CONFIG")
}

// chDriver is the ClickHouse engine behind the store, or nil when it is not
// configured: EnableTiering and Exec exist only on the driver, not on the
// store interfaces.
func (r *chStoreRepository) chDriver() *clickhouse.Driver {
	d, _ := r.store.Store.(*clickhouse.Driver)
	return d
}

func toStore(r domain.Retention) store.Retention {
	return store.Retention{Keep: r.Keep(), ColdAfter: r.Cold()}
}

func days(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int(d / domain.Day)
}
