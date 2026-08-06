package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	purgeEvery      = time.Hour
	purgeBatch      = 10_000
	purgeMaxBatches = 100
	purgeTimeout    = 2 * time.Minute
)

// Leases is how the replicas agree that one of them purges. Without it every
// replica deletes the same rows on the same tick: not wrong, but N times the
// work and N writers contending for the oldest pages.
type Leases interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

type purger struct {
	repo   connectors.Repository
	leases Leases
	retain time.Duration
}

func newPurger(repo connectors.Repository, leases Leases, retainDays int) *purger {
	if retainDays <= 0 {
		return nil
	}
	return &purger{repo: repo, leases: leases, retain: time.Duration(retainDays) * 24 * time.Hour}
}

func (p *purger) Start(ctx context.Context) {
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

const purgeLease = "audit-purge"

func (p *purger) runOnce(ctx context.Context) {
	if p.leases != nil {
		mine, err := p.leases.Acquire(ctx, purgeLease, purgeEvery)
		if err != nil {
			_ = catcher.Error("audit purge: cannot take the lease", err, nil)
			return
		}
		if !mine {
			return
		}
	}

	cutoff := time.Now().UTC().Add(-p.retain)

	var total int64
	for range purgeMaxBatches {
		rCtx, cancel := context.WithTimeout(ctx, purgeTimeout)
		n, err := p.repo.DeleteOlderThan(tenancy.WithAllTenants(rCtx), cutoff, purgeBatch)
		cancel()

		if err != nil {
			_ = catcher.Error("audit purge failed", err, map[string]any{"removed": total})
			return
		}
		total += n
		if n < purgeBatch {
			break
		}
	}

	if total > 0 {
		catcher.Info("audit entries past retention removed", map[string]any{
			"removed": total,
			"before":  cutoff.Format(time.RFC3339),
		})
	}
}
