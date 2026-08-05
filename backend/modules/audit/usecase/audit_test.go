package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const customerTenant = "8f1c1b8e-0000-4000-8000-000000000001"

type fakeRepo struct {
	mu       sync.Mutex
	rows     []*domain.AuditLog
	batches  int
	purged   int64
	spanning bool
}

func (f *fakeRepo) InsertBatch(ctx context.Context, rows []*domain.AuditLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.batches++
	f.spanning = tenancy.SpansAllTenants(ctx)
	f.rows = append(f.rows, rows...)
	return nil
}

func (f *fakeRepo) List(context.Context, common_models.IListRequest) (common_models.ListResponse[domain.AuditLog], error) {
	return common_models.ListResponse[domain.AuditLog]{}, nil
}

func (f *fakeRepo) GetByID(context.Context, uint64) (*domain.AuditLog, error) { return nil, nil }

func (f *fakeRepo) DeleteOlderThan(_ context.Context, cutoff time.Time, limit int) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	kept := make([]*domain.AuditLog, 0, len(f.rows))
	var removed int64
	for _, r := range f.rows {
		if r.Timestamp.Before(cutoff) && removed < int64(limit) {
			removed++
			continue
		}
		kept = append(kept, r)
	}
	f.rows = kept
	f.purged += removed
	return removed, nil
}

func (f *fakeRepo) snapshot() ([]*domain.AuditLog, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*domain.AuditLog, len(f.rows))
	copy(out, f.rows)
	return out, f.batches
}

func newService(t *testing.T) (*Service, *fakeRepo) {
	t.Helper()
	repo := &fakeRepo{}
	s := New(repo, 0)
	s.Start(context.Background())
	t.Cleanup(s.Stop)
	return s, repo
}

// Stop is what makes the queue safe: everything recorded before it must be
// written, since it runs after the HTTP server has drained.
func TestStopWritesWhatIsQueued(t *testing.T) {
	repo := &fakeRepo{}
	s := New(repo, 0)
	s.Start(context.Background())

	ctx := authz.WithTenantID(context.Background(), customerTenant)
	for range 50 {
		s.Log(ctx, connectors.Event{Action: "user.update"})
	}
	s.Stop()

	rows, _ := repo.snapshot()
	if len(rows) != 50 {
		t.Fatalf("wrote %d entries, want 50", len(rows))
	}
}

// A batch holds rows from whichever tenants were active, so the write cannot be
// scoped to one — each row carries its own.
func TestBatchesAreWrittenUnscoped(t *testing.T) {
	repo := &fakeRepo{}
	s := New(repo, 0)
	s.Start(context.Background())
	s.Log(authz.WithTenantID(context.Background(), customerTenant), connectors.Event{Action: "x"})
	s.Stop()

	if !repo.spanning {
		t.Error("the batch insert was scoped to a tenant")
	}
}

func TestTenantComesFromTheContext(t *testing.T) {
	s, repo := newService(t)

	s.Log(authz.WithTenantID(context.Background(), customerTenant), connectors.Event{Action: "user.update"})
	s.Stop()

	rows, _ := repo.snapshot()
	if len(rows) != 1 || rows[0].TenantID != customerTenant {
		t.Fatalf("rows = %+v, want one in %s", rows, customerTenant)
	}
}

// An entry with no tenant is invisible the day an instance takes an MSSP
// licence, and the migration put everything that predates tenancy in the
// default tenant.
func TestNoTenantFallsBackToTheDefault(t *testing.T) {
	s, repo := newService(t)

	s.Log(context.Background(), connectors.Event{Action: "auth.login"})
	s.Stop()

	rows, _ := repo.snapshot()
	if len(rows) != 1 || rows[0].TenantID != authz.DefaultTenantID {
		t.Fatalf("rows = %+v, want one in the default tenant", rows)
	}
}

// The entry lands in the tenant that was looked at, carrying the platform
// administrator who did it and under what grant.
func TestSupportSessionIsRecorded(t *testing.T) {
	s, repo := newService(t)

	ctx := authz.WithTenantID(context.Background(), customerTenant)
	s.Log(ctx, connectors.Event{
		Action:        "alert.read",
		UserLogin:     "admin",
		SupportAccess: authz.SupportRead,
	})
	s.Stop()

	rows, _ := repo.snapshot()
	if len(rows) != 1 {
		t.Fatalf("wrote %d entries, want 1", len(rows))
	}
	if rows[0].TenantID != customerTenant {
		t.Errorf("tenant = %q, want the tenant that was supported", rows[0].TenantID)
	}
	if rows[0].SupportAccess != authz.SupportRead {
		t.Errorf("supportAccess = %q, want READ", rows[0].SupportAccess)
	}
	if rows[0].UserLogin != "admin" {
		t.Errorf("userLogin = %q, want the platform administrator", rows[0].UserLogin)
	}
}

// A full queue writes synchronously rather than dropping. Losing the trail is
// the one outcome that is not acceptable.
func TestAFullQueueFallsBackToWritingInline(t *testing.T) {
	repo := &fakeRepo{}
	s := New(repo, 0) // deliberately not started: nothing drains

	for range queueSize + 10 {
		s.Log(context.Background(), connectors.Event{Action: "user.update"})
	}

	// The 10 that did not fit were written where Log was called; the rest are
	// still in the queue. Nothing was dropped.
	rows, batches := repo.snapshot()
	if len(rows) != 10 {
		t.Errorf("rows written = %d, want the 10 that did not fit", len(rows))
	}
	if batches != 10 {
		t.Errorf("writes = %d, want one per entry that overflowed", batches)
	}
}

// Ordinary traffic is written in batches, not one row per request: that is the
// whole point of taking the insert off the request path.
func TestTrafficIsBatched(t *testing.T) {
	repo := &fakeRepo{}
	s := New(repo, 0)
	s.Start(context.Background())

	for range batchMax {
		s.Log(context.Background(), connectors.Event{Action: "user.update"})
	}

	deadline := time.After(2 * time.Second)
	for {
		rows, batches := repo.snapshot()
		if len(rows) == batchMax {
			if batches > 2 {
				t.Fatalf("writes = %d for %d entries; want them batched", batches, batchMax)
			}
			break
		}
		select {
		case <-deadline:
			t.Fatalf("only %d of %d entries were written", len(rows), batchMax)
		default:
		}
	}
	s.Stop()
}

func aged(days int) *domain.AuditLog {
	return &domain.AuditLog{Timestamp: time.Now().UTC().AddDate(0, 0, -days)}
}

func seed(repo *fakeRepo, rows ...*domain.AuditLog) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.rows = append(repo.rows, rows...)
}

func TestPurgeRemovesOnlyWhatIsPastRetention(t *testing.T) {
	repo := &fakeRepo{}
	seed(repo, aged(400), aged(370), aged(360), aged(1))

	newPurger(repo, 365).runOnce(context.Background())

	rows, _ := repo.snapshot()
	if len(rows) != 2 {
		t.Fatalf("kept %d entries, want the 2 inside the year", len(rows))
	}
	for _, r := range rows {
		if time.Since(r.Timestamp) > 365*24*time.Hour {
			t.Errorf("kept an entry from %s, which is past retention", r.Timestamp)
		}
	}
}

// No retention means keep everything: an instance whose obligations say never
// to delete must not have entries removed by a default.
func TestNoRetentionKeepsEverything(t *testing.T) {
	repo := &fakeRepo{}
	seed(repo, aged(4000))

	p := newPurger(repo, 0)
	if p != nil {
		t.Fatal("a retention of 0 built a purger")
	}
	p.Start(context.Background()) // nil receiver, must not panic

	rows, _ := repo.snapshot()
	if len(rows) != 1 {
		t.Errorf("kept %d entries, want everything", len(rows))
	}
}

// One run is bounded so the first purge after a retention is set does not hold
// the database down until years of backlog are gone.
func TestPurgeIsBoundedPerRun(t *testing.T) {
	repo := &fakeRepo{}
	old := make([]*domain.AuditLog, 0, purgeBatch*purgeMaxBatches+500)
	for range cap(old) {
		old = append(old, aged(400))
	}
	seed(repo, old...)

	newPurger(repo, 365).runOnce(context.Background())

	rows, _ := repo.snapshot()
	if len(rows) != 500 {
		t.Fatalf("left %d entries, want the 500 beyond one run's cap", len(rows))
	}
}
