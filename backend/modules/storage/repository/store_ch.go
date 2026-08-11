package repository

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/store"

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
	domain.DatasetLogs:   eventstore.DatasetLogs,
	domain.DatasetAlerts: eventstore.DatasetAlerts,
	domain.DatasetStats:  eventstore.DatasetStats,
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
	return r.store.EnableTiering(ctx, ds, domain.PolicyName, toStore(want))
}

func (r *chStoreRepository) Usage(ctx context.Context) ([]domain.Usage, error) {
	rows, err := r.store.Usage(ctx)
	if err != nil {
		return nil, err
	}

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
	var oldest, newest time.Time
	err := r.store.Conn.QueryRow(ctx,
		"SELECT min(min_time), max(max_time) FROM system.parts WHERE database = currentDatabase() AND table = ? AND active",
		string(d)).Scan(&oldest, &newest)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return oldest.UTC(), newest.UTC(), nil
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
	var n uint64
	err := r.store.Conn.QueryRow(ctx,
		"SELECT count() FROM system.storage_policies WHERE policy_name = ? AND volume_name = ?",
		domain.PolicyName, domain.ColdVolume).Scan(&n)
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *chStoreRepository) ReloadConfig(ctx context.Context) error {
	return r.store.Conn.Exec(ctx, "SYSTEM RELOAD CONFIG")
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
