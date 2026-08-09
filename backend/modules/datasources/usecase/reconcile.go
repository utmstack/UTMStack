package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/tenancy"

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

const leaseName = "datasources.reconcile"

func (s *StatsReconciler) runLogged(ctx context.Context) {
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
			Name:         src.DataSource,
			DataType:     src.DataType,
			SourceKind:   domain.SourceKindDirect,
			LastPingAt:   &last,
			ModifiedAt:   &now,
			DiscoveredAt: &now,
		})
	}
	return s.repo.UpsertLivenessBatch(tenancy.WithAllTenants(ctx), items)
}
