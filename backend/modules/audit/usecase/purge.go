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

type purger struct {
	repo   connectors.Repository
	retain time.Duration
}

func newPurger(repo connectors.Repository, retainDays int) *purger {
	if retainDays <= 0 {
		return nil
	}
	return &purger{repo: repo, retain: time.Duration(retainDays) * 24 * time.Hour}
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

func (p *purger) runOnce(ctx context.Context) {
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
