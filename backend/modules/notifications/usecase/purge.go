package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/notifications/connectors"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	purgeEvery      = 6 * time.Hour
	purgeBatch      = 1000
	purgeMaxBatches = 50
	purgeTimeout    = 30 * time.Second
	purgeLease      = "notifications-purge"
)

// Leases is how the replicas agree that one of them purges.
type Leases interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

// Purger keeps the table from growing without end. A read notification has
// already done its job, so it goes early; an unread one is still information
// somebody has not seen, and only leaves when it is old enough to be useless.
type Purger struct {
	repo      connectors.NotificationRepository
	leases    Leases
	readAfter time.Duration
	afterAll  time.Duration
}

func NewPurger(repo connectors.NotificationRepository, leases Leases, readDays, allDays int) *Purger {
	if readDays <= 0 && allDays <= 0 {
		return nil
	}
	return &Purger{
		repo:      repo,
		leases:    leases,
		readAfter: time.Duration(readDays) * 24 * time.Hour,
		afterAll:  time.Duration(allDays) * 24 * time.Hour,
	}
}

func (p *Purger) Start(ctx context.Context) {
	if p == nil {
		return
	}
	t := time.NewTicker(purgeEvery)
	defer t.Stop()

	p.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			p.runOnce(ctx)
		}
	}
}

func (p *Purger) runOnce(ctx context.Context) {
	if p.leases != nil {
		mine, err := p.leases.Acquire(ctx, purgeLease, purgeEvery)
		if err != nil {
			_ = catcher.Error("notifications purge: cannot take the lease", err, nil)
			return
		}
		if !mine {
			return
		}
	}

	now := time.Now().UTC()
	if p.readAfter > 0 {
		p.sweep(ctx, now.Add(-p.readAfter), true)
	}
	if p.afterAll > 0 {
		p.sweep(ctx, now.Add(-p.afterAll), false)
	}
}

// sweep spans every tenant: retention is the instance's rule, and a tenant that
// never signs in still must not keep its rows for ever.
func (p *Purger) sweep(ctx context.Context, cutoff time.Time, onlyRead bool) {
	var total int64
	for range purgeMaxBatches {
		rCtx, cancel := context.WithTimeout(ctx, purgeTimeout)
		n, err := p.repo.DeleteOlderThan(tenancy.WithAllTenants(rCtx), cutoff, onlyRead, purgeBatch)
		cancel()
		if err != nil {
			_ = catcher.Error("notifications purge failed", err, map[string]any{"removed": total, "onlyRead": onlyRead})
			return
		}
		total += n
		if n < purgeBatch {
			break
		}
	}
}
