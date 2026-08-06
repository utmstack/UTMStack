package usecase

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeLeases struct {
	mine  bool
	err   error
	calls int
}

func (f *fakeLeases) Acquire(context.Context, string, time.Duration) (bool, error) {
	f.calls++
	return f.mine, f.err
}

// The service is built without a store on purpose: reaching EnsureDefault at
// all is the failure these cases are looking for, so a nil store would panic
// loudly rather than quietly pass.
func TestEnsureDefaultIfMineSkipsWhenAnotherReplicaHoldsTheLease(t *testing.T) {
	leases := &fakeLeases{mine: false}
	s := &ConfigService{leases: leases}

	if s.ensureDefaultIfMine() {
		t.Fatal("reported done while another replica holds the lease")
	}
	if leases.calls != 1 {
		t.Fatalf("asked for the lease %d times, want 1", leases.calls)
	}
}

// A lease that cannot be read is not permission to run: two replicas that both
// fail to reach the database would otherwise both provision.
func TestEnsureDefaultIfMineSkipsWhenTheLeaseCannotBeRead(t *testing.T) {
	s := &ConfigService{leases: &fakeLeases{err: errors.New("database is down")}}

	if s.ensureDefaultIfMine() {
		t.Fatal("ran the work despite failing to take the lease")
	}
}
