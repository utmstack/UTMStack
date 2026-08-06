// Package joblease lets one replica run a periodic job.
//
// Some background work has a row to claim — SOAR claims the execution it is
// about to dispatch, so the others skip it. Work that is a whole pass over
// something has no such row: every replica would do all of it. A lease is the
// row those jobs claim instead.
//
// It is a lease rather than a lock so a replica that dies does not hold it
// forever: whoever asks after it expires takes it.
package joblease

import (
	"context"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Lease struct {
	Name      string    `gorm:"column:name;primaryKey;size:64"`
	Holder    string    `gorm:"column:holder;size:128;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
}

func (Lease) TableName() string { return "job_leases" }

type Leases struct {
	db     *gorm.DB
	holder string
}

func New(db *gorm.DB) *Leases {
	host, _ := os.Hostname()
	return &Leases{db: db, holder: fmt.Sprintf("%s/%d", host, os.Getpid())}
}

// Acquire reports whether this replica should do the work now.
//
// The whole decision is one statement, which is what makes it safe: the insert
// takes a lease nobody holds, and the conflict path takes one that has expired.
// A replica whose lease is still live matches neither, so exactly one caller
// per period comes away with it.
//
// ttl should exceed how long the job takes. Too short and a second replica
// starts while the first is still running; too long and a crash leaves the work
// undone until it expires.
func (l *Leases) Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error) {
	if l == nil {
		return true, nil // no database to coordinate through: run it
	}

	now := time.Now().UTC()
	res := l.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"holder":     l.holder,
				"expires_at": now.Add(ttl),
			}),
			Where: clause.Where{Exprs: []clause.Expression{
				gorm.Expr("job_leases.expires_at < ?", now),
			}},
		}).
		Create(&Lease{Name: name, Holder: l.holder, ExpiresAt: now.Add(ttl)})

	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}
