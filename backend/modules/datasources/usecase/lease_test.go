package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/datasources/connectors"
)

type fakeLeases struct {
	mine  bool
	asked int
}

func (f *fakeLeases) Acquire(context.Context, string, time.Duration) (bool, error) {
	f.asked++
	return f.mine, nil
}

// Every replica ticks and the pass is the same work for all of them. Only the
// one holding the lease does it; the rest skip without touching the store.
func TestOnlyTheLeaseHolderReconciles(t *testing.T) {
	stats := &fakeStats{rows: []connectors.StatSource{
		{TenantID: tenantA, DataSource: "fw-01", DataType: "syslog", LastSeen: time.Now().UTC()},
	}}

	t.Run("holder does the work", func(t *testing.T) {
		repo := &fakeRepo{}
		r := NewStatsReconciler(repo, stats, &fakeLeases{mine: true})
		r.runLogged(context.Background())
		if len(repo.got) == 0 {
			t.Error("the holder did not reconcile")
		}
	})

	t.Run("the others skip", func(t *testing.T) {
		repo := &fakeRepo{}
		leases := &fakeLeases{mine: false}
		r := NewStatsReconciler(repo, stats, leases)
		r.runLogged(context.Background())
		if leases.asked != 1 {
			t.Errorf("asked for the lease %d times, want 1", leases.asked)
		}
		if len(repo.got) != 0 {
			t.Error("a replica without the lease wrote anyway")
		}
	})
}

// Without a way to coordinate, every replica runs — which is what it did before
// leases existed, and is wasteful rather than wrong.
func TestWithoutLeasesEveryReplicaRuns(t *testing.T) {
	stats := &fakeStats{rows: []connectors.StatSource{
		{TenantID: tenantA, DataSource: "fw-01", DataType: "syslog", LastSeen: time.Now().UTC()},
	}}
	repo := &fakeRepo{}

	NewStatsReconciler(repo, stats, nil).runLogged(context.Background())

	if len(repo.got) == 0 {
		t.Error("nothing ran without a lease provider")
	}
}
