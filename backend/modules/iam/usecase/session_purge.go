package usecase

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	sessionPurgeEvery = 6 * time.Hour
	sessionPurgeBatch = 1000
	sessionPurgeMax   = 50
	sessionPurgeLease = "session-purge"

	// revokedGrace keeps a closed session readable for a while. After an
	// incident the question is when a session ended, and deleting the row on
	// revocation is what makes that unanswerable.
	revokedGrace = 7 * 24 * time.Hour
)

type SessionLeases interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

type SessionPurger struct {
	repo   connectors.RefreshTokenRepository
	leases SessionLeases
}

func NewSessionPurger(repo connectors.RefreshTokenRepository, leases SessionLeases) *SessionPurger {
	return &SessionPurger{repo: repo, leases: leases}
}

func (p *SessionPurger) Start(ctx context.Context) {
	if p == nil {
		return
	}
	t := time.NewTicker(sessionPurgeEvery)
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

func (p *SessionPurger) runOnce(ctx context.Context) {
	if p.leases != nil {
		mine, err := p.leases.Acquire(ctx, sessionPurgeLease, sessionPurgeEvery)
		if err != nil {
			_ = catcher.Error("session purge: cannot take the lease", err, nil)
			return
		}
		if !mine {
			return
		}
	}

	now := time.Now().UTC()
	var total int64
	for range sessionPurgeMax {
		n, err := p.repo.DeleteSpent(tenancy.WithAllTenants(ctx), now.Add(-revokedGrace), now, sessionPurgeBatch)
		if err != nil {
			_ = catcher.Error("session purge failed", err, map[string]any{"removed": total})
			return
		}
		total += n
		if n < sessionPurgeBatch {
			break
		}
	}
}
