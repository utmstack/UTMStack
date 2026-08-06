package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
	"github.com/utmstack/utmstack/backend/modules/datasources/domain"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	tenantA = "8f1c1b8e-0000-4000-8000-00000000000a"
	tenantB = "8f1c1b8e-0000-4000-8000-00000000000b"
)

type fakeStats struct{ rows []connectors.StatSource }

func (f *fakeStats) DistinctSources(context.Context, time.Time, time.Time) ([]connectors.StatSource, error) {
	return f.rows, nil
}

type fakeRepo struct {
	connectors.DatasourceRepository
	got      []domain.Datasource
	spanning bool
}

func (f *fakeRepo) UpsertLivenessBatch(ctx context.Context, items []domain.Datasource) error {
	f.spanning = tenancy.SpansAllTenants(ctx)
	f.got = append(f.got, items...)
	return nil
}

// One reconciler serves every tenant, so what it registers spans them. Each row
// carries its own tenant; scoped to the caller's, every tenant's direct sources
// would collapse onto whichever tenant it happened to run as — and it runs as
// none, which under a multi-tenant licence is an error rather than a default.
func TestEachSourceKeepsItsOwnTenant(t *testing.T) {
	seen := time.Now().UTC().Add(-3 * time.Hour)
	stats := &fakeStats{rows: []connectors.StatSource{
		{TenantID: tenantA, DataSource: "fw-01", DataType: "syslog", LastSeen: seen},
		{TenantID: tenantB, DataSource: "fw-01", DataType: "syslog", LastSeen: seen},
	}}
	repo := &fakeRepo{}

	if err := NewStatsReconciler(repo, stats, nil).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !repo.spanning {
		t.Error("the write was scoped to one tenant")
	}
	if len(repo.got) != 2 {
		t.Fatalf("registered %d sources, want 2", len(repo.got))
	}
	// The same host name in two tenants is two datasources, not one.
	if repo.got[0].TenantID == repo.got[1].TenantID {
		t.Errorf("both landed in %s", repo.got[0].TenantID)
	}
	for _, d := range repo.got {
		if d.TenantID == "" {
			t.Error("a source was registered with no tenant")
		}
	}
}

// The window is a day, so the time a source was actually last seen is what
// makes liveness mean anything: stamping now would report one that stopped this
// morning as live.
func TestTheLastSeenIsWhatTheStatisticsSaid(t *testing.T) {
	seen := time.Now().UTC().Add(-20 * time.Hour)
	stats := &fakeStats{rows: []connectors.StatSource{
		{TenantID: tenantA, DataSource: "fw-01", DataType: "syslog", LastSeen: seen},
	}}
	repo := &fakeRepo{}

	if err := NewStatsReconciler(repo, stats, nil).Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := repo.got[0].LastPingAt; got == nil || !got.Equal(seen) {
		t.Errorf("lastPingAt = %v, want %v", got, seen)
	}
}
