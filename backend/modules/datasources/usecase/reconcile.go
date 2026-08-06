package usecase

import (
	"context"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
)

const (
	reconcileWindow   = 24 * time.Hour
	reconcileInterval = 10 * time.Minute
)

type leaser interface {
	Acquire(ctx context.Context, name string, ttl time.Duration) (bool, error)
}

type StatsReconciler struct {
	repo   connectors.DatasourceRepository
	stats  connectors.StatsReader
	leases leaser
}

// NewStatsReconciler takes leases nil when there is nothing to coordinate
// through, and every replica then runs every pass — which is what it did before
// and is wasteful rather than wrong.
func NewStatsReconciler(repo connectors.DatasourceRepository, stats connectors.StatsReader, leases leaser) *StatsReconciler {
	return &StatsReconciler{repo: repo, stats: stats, leases: leases}
}

func (s *StatsReconciler) Start(ctx context.Context) {
	s.runLogged(ctx)
	t := time.NewTicker(reconcileInterval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.runLogged(ctx)
		}
	}
}

// leaseName is the work, not the worker: whichever replica holds it does this
// pass and the rest skip it.
const leaseName = "datasources.reconcile"

func (s *StatsReconciler) runLogged(ctx context.Context) {
	// Every replica ticks, and the pass is the same query and the same upsert
	// for all of them. Idempotent, so N replicas corrupt nothing — they just do
	// the work N times.
	if s.leases != nil {
		mine, err := s.leases.Acquire(ctx, leaseName, reconcileInterval)
		if err != nil {
			_ = catcher.Error("datasources: could not take the reconcile lease", err, nil)
			return
		}
		if !mine {
			return
		}
	}

	if err := s.Run(ctx); err != nil {
		_ = catcher.Error("datasources: stats reconcile failed", err, nil)
	}
}

// Run registers what ingested since the last pass. It reads across tenants and
// writes across them, which is why nothing here reads a tenant from the context:
// the reconciler belongs to no tenant, and every row says whose it is.
func (s *StatsReconciler) Run(ctx context.Context) error {
	now := time.Now().UTC()
	sources, err := s.stats.DistinctSources(ctx, now.Add(-reconcileWindow), now)
	if err != nil {
		return err
	}
	if len(sources) == 0 {
		return nil
	}

	items := make([]domain.Datasource, 0, len(sources))
	for _, src := range sources {
		last := src.LastSeen
		if last.IsZero() {
			last = now
		}
		items = append(items, domain.Datasource{
			TenantID:     src.TenantID,
			SourceRef:    src.DataType + ":" + src.DataSource,
			Name:         src.DataSource,
			DataType:     src.DataType,
			SourceKind:   "direct",
			LastPingAt:   &last,
			ModifiedAt:   &now,
			DiscoveredAt: &now,
		})
	}
	// Unscoped: the batch spans every tenant that ingested, and each row carries
	// its own. Scoped to the caller's, this would collapse them onto whichever
	// tenant the reconciler happens to run as — and it runs as none.
	return s.repo.UpsertLivenessBatch(tenancy.WithAllTenants(ctx), items)
}
