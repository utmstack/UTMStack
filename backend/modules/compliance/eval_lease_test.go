package compliance

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeLeases mimics joblease semantics: first Acquire for a name wins until
// the ttl passes; overlapping Acquires from other "holders" return false.
type fakeLeases struct {
	mu       sync.Mutex
	held     map[string]time.Time // name → expiresAt
	holder   string
	captured []string
}

func (f *fakeLeases) Acquire(_ context.Context, name string, ttl time.Duration) (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.captured = append(f.captured, name)
	if exp, ok := f.held[name]; ok && exp.After(time.Now()) {
		return false, nil
	}
	if f.held == nil {
		f.held = map[string]time.Time{}
	}
	f.held[name] = time.Now().Add(ttl)
	return true, nil
}

// TestEvalLeaseNameFitsColumn: joblease.Lease.Name is size:64. The pattern
// c:e:<uuid-no-dashes>:<framework_key> must stay under it for real inputs.
func TestEvalLeaseNameFitsColumn(t *testing.T) {
	tid := "8f1c1b8e-0000-4000-8000-000000000001" // 36 chars, dashes
	// The longest framework key that ships is nist-800-171-r3-ish; give
	// ourselves 20 chars of headroom.
	key := strings.Repeat("k", 20)
	name := evalLeaseName(tid, key)
	if len(name) > 64 {
		t.Fatalf("lease name overflowed job_leases.name(64): %d chars: %q", len(name), name)
	}
	if strings.Contains(name, "-") {
		t.Fatalf("lease name still contains a dash from the uuid: %q", name)
	}
}

// TestOneReplicaWinsPerPair: two acquisitions of the same lease within the
// TTL window — exactly one comes back as mine=true, the other skips.
func TestOneReplicaWinsPerPair(t *testing.T) {
	leases := &fakeLeases{}
	// Module set up just enough for claimEval; leases is the only field it
	// touches on this path.
	m := &Module{leases: leases}
	ctx := context.Background()

	first := m.claimEval(ctx, "tA", "hipaa")
	second := m.claimEval(ctx, "tA", "hipaa")

	if !first {
		t.Fatal("first replica should have won the lease")
	}
	if second {
		t.Fatal("second replica should have lost — the lease was still live")
	}
	// Different (tenant, framework) shouldn't collide.
	if !m.claimEval(ctx, "tB", "hipaa") {
		t.Fatal("different tenant shouldn't be blocked by tA's lease")
	}
	if !m.claimEval(ctx, "tA", "iso") {
		t.Fatal("different framework shouldn't be blocked by hipaa's lease")
	}
}

// A nil leases means single-instance / no coordination — every call runs.
func TestNilLeasesRunsUncoordinated(t *testing.T) {
	m := &Module{}
	if !m.claimEval(context.Background(), "any", "any") {
		t.Fatal("nil leases must return true so single-instance keeps running")
	}
}
