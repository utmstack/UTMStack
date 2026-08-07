package repository

import (
	"context"
	"database/sql"
	"hash/fnv"
	"sort"

	"gorm.io/gorm"
)

type alertLocker struct {
	db *sql.DB
}

func newAlertLocker(gdb *gorm.DB) *alertLocker {
	if gdb == nil {
		return &alertLocker{}
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return &alertLocker{}
	}
	return &alertLocker{db: sqlDB}
}

func (l *alertLocker) lock(ctx context.Context, alertIDs []string) (release func(), err error) {
	if l.db == nil || len(alertIDs) == 0 {
		return func() {}, nil
	}

	keys := lockKeys(alertIDs)

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx,
		"SELECT pg_advisory_xact_lock(k) FROM unnest($1::bigint[]) AS k", keys); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	return func() { _ = tx.Rollback() }, nil
}

func lockKeys(alertIDs []string) []int64 {
	seen := make(map[int64]struct{}, len(alertIDs))
	keys := make([]int64, 0, len(alertIDs))
	for _, id := range alertIDs {
		h := fnv.New64a()
		_, _ = h.Write([]byte(id))
		k := int64(h.Sum64())
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
