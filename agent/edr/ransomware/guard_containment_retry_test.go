package ransomware

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type retryResponse struct {
	mu    sync.Mutex
	calls int
	run   func(int, int64, int) ([]int, error)
}

func (r *retryResponse) KillTree(pid int, _ string) ([]int, error) { return r.kill(pid, 0) }
func (r *retryResponse) KillTreeIdentity(pid int, id int64, _ string) ([]int, error) {
	return r.kill(pid, id)
}
func (r *retryResponse) kill(pid int, id int64) ([]int, error) {
	r.mu.Lock()
	r.calls++
	n := r.calls
	r.mu.Unlock()
	return r.run(pid, id, n)
}
func (r *retryResponse) Suspend(int) error                { return nil }
func (r *retryResponse) Resume(int) error                 { return nil }
func (r *retryResponse) SuspendIdentity(int, int64) error { return nil }
func (r *retryResponse) ResumeIdentity(int, int64) error  { return nil }
func (r *retryResponse) count() int                       { r.mu.Lock(); defer r.mu.Unlock(); return r.calls }

type retryEvidence struct {
	mu        sync.Mutex
	incidents []cache.RansomwareIncident
}

func (*retryEvidence) Append(string) error { return nil }
func (e *retryEvidence) StoreIncident(r cache.RansomwareIncident) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.incidents = append(e.incidents, r)
	return nil
}
func (e *retryEvidence) actions() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var out []string
	for _, r := range e.incidents {
		out = append(out, r.Action)
	}
	return out
}

func retryGuard(run func(int, int64, int) ([]int, error)) (*Guard, *retryResponse, *retryEvidence, FileEvent) {
	g, _, _, _, _, table := newTestGuard("kill", map[string]bool{"/scope/canary": true}, 50, 100)
	r := &retryResponse{run: run}
	evidence := &retryEvidence{}
	g.deps.Resp, g.deps.Spool, g.deps.Incidents = r, evidence, evidence
	table.Add(proctable.Proc{PID: 700, Image: "/owned/actor", OSIdentity: 123, IdentityRequired: true})
	return g, r, evidence, FileEvent{PID: 700, Path: "/scope/canary", Op: OpWrite, OSIdentity: 123, IdentityRequired: true}
}
func TestGuardFailedContainmentRetriesFreshCanary(t *testing.T) {
	for _, tc := range []struct {
		name   string
		killed []int
		err    error
	}{
		{"transient_error", nil, errors.New("temporary response error")},
		{"no_confirmed_kill", nil, nil},
		{"child_only", []int{701}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, r, e, fe := retryGuard(func(pid int, _ int64, n int) ([]int, error) {
				if n == 1 {
					return tc.killed, tc.err
				}
				return []int{pid}, nil
			})
			g.OnFileEvent(fe)
			g.OnFileEvent(fe)
			g.OnFileEvent(fe)
			if r.count() != 2 {
				t.Fatalf("fresh canary after failed root response: calls=%d, want 2", r.count())
			}
			actions := e.actions()
			if len(actions) != 2 || actions[0] != "contain_failed" || actions[1] != "contained" {
				t.Fatalf("response evidence %v", actions)
			}
		})
	}
}
func TestGuardConfirmedRootKillRetainsDedupOnOtherFailure(t *testing.T) {
	g, r, _, fe := retryGuard(func(pid int, _ int64, _ int) ([]int, error) { return []int{pid}, errors.New("child response failed") })
	g.OnFileEvent(fe)
	g.OnFileEvent(fe)
	if r.count() != 1 {
		t.Fatalf("confirmed root kill retried due to child error: %d", r.count())
	}
	g2, legacy, _, q, _, table := newTestGuard("kill", map[string]bool{"/scope/canary": true}, 50, 100)
	q.err = errors.New("quarantine unavailable")
	table.Add(proctable.Proc{PID: 710, Image: "/actor", StartTS: 1})
	g2.OnFileEvent(FileEvent{PID: 710, Path: "/scope/canary", Op: OpWrite})
	g2.OnFileEvent(FileEvent{PID: 710, Path: "/scope/canary", Op: OpWrite})
	if legacy.killCount() != 1 {
		t.Fatal("confirmed root kill retried due to quarantine error")
	}
}
func TestGuardFailedContainmentPreservesConcurrentDedup(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})
	g, r, _, fe := retryGuard(func(pid int, _ int64, n int) ([]int, error) {
		if n == 1 {
			close(entered)
			<-release
			return nil, errors.New("temporary")
		}
		return []int{pid}, nil
	})
	var now atomic.Int64
	now.Store(1000)
	clock := func() time.Time { return time.Unix(now.Load(), 0) }
	g.deps.Now, g.scorer.now = clock, clock
	done := make(chan struct{})
	go func() { defer close(done); g.OnFileEvent(fe) }()
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("first response did not start")
	}
	// An in-flight response stays reserved even beyond the completed-mark TTL.
	now.Store(1000 + int64(containedTTL/time.Second) + 1)
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); g.OnFileEvent(fe) }()
	}
	wg.Wait()
	if r.count() != 1 {
		t.Fatal("concurrent evidence duplicated in-flight response")
	}
	close(release)
	<-done
	g.OnFileEvent(fe)
	if r.count() != 2 {
		t.Fatalf("fresh evidence could not retry completed failure: %d", r.count())
	}
}
func TestGuardOldFailedResponseCannotEraseNewIdentityClaim(t *testing.T) {
	oldEntered, newEntered, oldRelease, newRelease := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	g, r, _, fe := retryGuard(func(pid int, id int64, _ int) ([]int, error) {
		switch id {
		case 123:
			close(oldEntered)
			<-oldRelease
			return nil, errors.New("old failed")
		case 124:
			close(newEntered)
			<-newRelease
			return []int{pid}, nil
		}
		t.Errorf("unexpected identity %d", id)
		return nil, nil
	})
	oldDone := make(chan struct{})
	go func() { defer close(oldDone); g.OnFileEvent(fe) }()
	<-oldEntered
	next := fe
	next.OSIdentity = 124
	g.deps.Table.Add(proctable.Proc{PID: fe.PID, Image: "/new/actor", OSIdentity: 124, IdentityRequired: true})
	newDone := make(chan struct{})
	go func() { defer close(newDone); g.OnFileEvent(next) }()
	<-newEntered
	close(oldRelease)
	<-oldDone
	if !g.isContainedObservation(proctable.Proc{PID: next.PID, OSIdentity: next.OSIdentity}) {
		t.Error("old failure erased new in-flight response claim")
	}
	g.scorer.mu.Lock()
	state := g.scorer.states[next.PID]
	preserved := state != nil && state.gen == next.OSIdentity
	g.scorer.mu.Unlock()
	if !preserved {
		t.Error("old response erased newer identity's scoring state")
	}
	close(newRelease)
	<-newDone
	g.OnFileEvent(next)
	if r.count() != 2 {
		t.Fatalf("new successful response was duplicated: %d", r.count())
	}
}

func TestScorerForgetGenerationPreservesReusedPIDAndUnknownBoundary(t *testing.T) {
	now := time.Unix(1000, 0)
	s := NewScorer(50, 100, 0, func() time.Time { return now })
	add := func(gen int64, weight float64) Decision {
		return s.Add(Evidence{PID: 700, Gen: gen, Kind: KindCanary, Weight: weight, TS: now})
	}
	if d := add(124, 30); d.Escalation != EscNone {
		t.Fatal(d)
	}
	s.ForgetGeneration(700, 123)
	s.ForgetGeneration(700, 0)
	if d := add(124, 30); d.Escalation != EscSuspend {
		t.Fatalf("stale or unknown completion erased live evidence: %+v", d)
	}
	s.ForgetGeneration(700, 124)
	if d := add(124, 30); d.Escalation != EscNone {
		t.Fatalf("matching completion did not release its evidence: %+v", d)
	}
	s.Forget(700)
	add(0, 30)
	s.ForgetGeneration(700, 0)
	if d := add(0, 30); d.Escalation != EscNone {
		t.Fatalf("unknown completion failed to release unknown evidence: %+v", d)
	}
}

func TestGuardDescendantCompletionPreservesNewIdentityResponse(t *testing.T) {
	parentEntered, parentRelease, childEntered, childRelease := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	g, _, _, fe := retryGuard(func(pid int, _ int64, _ int) ([]int, error) {
		if pid == 700 {
			close(parentEntered)
			<-parentRelease
			return []int{701, 700}, nil
		}
		close(childEntered)
		<-childRelease
		return []int{701}, nil
	})
	g.deps.Table.Add(proctable.Proc{PID: 701, PPID: 700, OSIdentity: 200, StartTS: 200, IdentityRequired: true})
	parentDone := make(chan struct{})
	go func() { defer close(parentDone); g.OnFileEvent(fe) }()
	<-parentEntered
	g.deps.Table.Add(proctable.Proc{PID: 701, PPID: 1, OSIdentity: 201, StartTS: 201, IdentityRequired: true})
	child := fe
	child.PID = 701
	child.OSIdentity = 201
	childDone := make(chan struct{})
	go func() { defer close(childDone); g.OnFileEvent(child) }()
	<-childEntered
	close(parentRelease)
	<-parentDone
	g.cmu.Lock()
	mark := g.contained[701]
	g.cmu.Unlock()
	if mark.osIdentity != 201 || !mark.pending {
		t.Errorf("old parent completion overwrote newer child response: %+v", mark)
	}
	g.scorer.mu.Lock()
	state := g.scorer.states[701]
	preserved := state != nil && state.gen == 201
	g.scorer.mu.Unlock()
	if !preserved {
		t.Error("old parent completion erased newer child evidence")
	}
	close(childRelease)
	<-childDone
}
func TestGuardDescendantCompletionUsesCapturedIdentityOnly(t *testing.T) {
	g, r, _, fe := retryGuard(func(pid int, _ int64, _ int) ([]int, error) { return []int{pid, 701, 702, 703}, nil })
	g.deps.Table.Add(proctable.Proc{PID: 701, PPID: 700, OSIdentity: 200, StartTS: 200, IdentityRequired: true})
	g.deps.Table.Add(proctable.Proc{PID: 703, PPID: 700})
	g.OnFileEvent(fe)
	child := fe
	child.PID = 701
	child.OSIdentity = 200
	g.OnFileEvent(child)
	if r.count() != 1 {
		t.Fatal("confirmed captured child was not deduplicated")
	}
	g.cmu.Lock()
	mark, known := g.contained[701]
	_, unknown := g.contained[702]
	_, unidentified := g.contained[703]
	g.cmu.Unlock()
	if !known || mark.osIdentity != 200 {
		t.Fatalf("captured child identity lost: %+v", mark)
	}
	if unknown || unidentified {
		t.Fatal("uncaptured or unidentified child acquired numeric-PID containment authority")
	}
}

func TestGuardDescendantCompletionPreservesPendingClaimWithoutTable(t *testing.T) {
	g, _, _, _ := retryGuard(func(pid int, _ int64, _ int) ([]int, error) { return []int{pid}, nil })
	token, claimed := g.claimContainment(701, 201, 201)
	if !claimed {
		t.Fatal("new child reservation failed")
	}
	g.scorer.Add(Evidence{PID: 701, Gen: 201, Kind: KindCanary, Weight: 100, TS: g.deps.Now()})
	g.completeDescendant(proctable.Proc{PID: 701, OSIdentity: 200, StartTS: 200})
	g.cmu.Lock()
	mark := g.contained[701]
	g.cmu.Unlock()
	if !mark.pending || mark.token != token || mark.osIdentity != 201 {
		t.Fatalf("old completion overrode pending native identity without table: %+v", mark)
	}
	if d := g.scorer.Add(Evidence{PID: 701, Gen: 201, Kind: KindCanary, Weight: 100, TS: g.deps.Now()}); d.Escalation != EscNone {
		t.Fatalf("old completion forgot pending child's prior escalation: %+v", d)
	}
}
