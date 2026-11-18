package ransomware

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

type admissionFeedFunc func(context.Context, func(FileEvent)) error

func (f admissionFeedFunc) Run(ctx context.Context, sink func(FileEvent)) error { return f(ctx, sink) }

type savedFeedAttempt struct {
	ctx  context.Context
	sink func(FileEvent)
}

type admissionCanaries struct {
	calls atomic.Int32
	check func(string) bool
}

func (c *admissionCanaries) Contains(path string) bool {
	c.calls.Add(1)
	return c.check(path)
}

func nextFeedAttempt(t *testing.T, attempts <-chan savedFeedAttempt) savedFeedAttempt {
	t.Helper()
	select {
	case attempt := <-attempts:
		return attempt
	case <-time.After(2 * time.Second):
		t.Fatal("feed attempt did not start")
		return savedFeedAttempt{}
	}
}

func waitAdmissionRun(t *testing.T, done <-chan error) {
	t.Helper()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Run returned %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run waited on a retained callback after cancellation")
	}
}

func TestGuardFeedRetainedSinksCannotReactivateAfterRetryOrParentCancel(t *testing.T) {
	old := feedRetryBackoff
	feedRetryBackoff = time.Millisecond
	defer func() { feedRetryBackoff = old }()
	for _, failure := range []error{errors.New("native feed dropped"), nil} {
		t.Run(func() string {
			if failure == nil {
				return "unexpected clean return"
			}
			return "failed return"
		}(), func(t *testing.T) {
			g, resp, app, _, incidents, table := newTestGuard("kill", nil, 50, 100)
			canaries := &admissionCanaries{check: func(string) bool { return true }}
			g.deps.Canaries = canaries
			for _, pid := range []int{101, 102, 103} {
				table.Add(proctable.Proc{PID: pid, Image: "/owned/harmless-test-process"})
			}
			attempts := make(chan savedFeedAttempt, 2)
			calls := 0
			feed := admissionFeedFunc(func(ctx context.Context, sink func(FileEvent)) error {
				calls++
				attempts <- savedFeedAttempt{ctx, sink}
				if calls == 1 {
					return failure
				}
				<-ctx.Done()
				return ctx.Err()
			})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- g.Run(ctx, feed, 999, nil) }()
			first, current := nextFeedAttempt(t, attempts), nextFeedAttempt(t, attempts)
			if first.ctx == current.ctx || first.ctx.Err() == nil || current.ctx.Err() != nil || ctx.Err() != nil {
				t.Fatal("attempt lifetimes were not independent")
			}
			first.sink(FileEvent{PID: 101, Path: "/owned/canary", Op: OpWrite})
			if canaries.calls.Load() != 0 || resp.killCount() != 0 || len(app.lines) != 0 || len(incidents.recs) != 0 {
				t.Fatal("retired attempt admitted filtering, evidence or response")
			}
			current.sink(FileEvent{PID: 102, Path: "/owned/canary", Op: OpWrite})
			if resp.killCount() != 1 || len(incidents.recs) != 1 {
				t.Fatal("active attempt failed to deliver its ordinary canary response")
			}
			cancel()
			waitAdmissionRun(t, done)
			before := canaries.calls.Load()
			current.sink(FileEvent{PID: 103, Path: "/owned/canary", Op: OpWrite})
			first.sink(FileEvent{PID: 103, Path: "/owned/canary", Op: OpWrite})
			if current.ctx.Err() == nil || canaries.calls.Load() != before || resp.killCount() != 1 || len(incidents.recs) != 1 || g.FeedHealthy() {
				t.Fatal("cancelled attempt admitted new work or stayed healthy")
			}
		})
	}
}

func TestGuardFeedCancellationDuringConsiderCannotStartEvidence(t *testing.T) {
	g, resp, app, _, incidents, table := newTestGuard("kill", nil, 50, 100)
	table.Add(proctable.Proc{PID: 101, Image: "/owned/harmless-test-process"})
	entered, release, callbackDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var enteredOnce, releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	canaries := &admissionCanaries{check: func(string) bool {
		enteredOnce.Do(func() { close(entered) })
		<-release
		return true
	}}
	g.deps.Canaries = canaries
	feed := admissionFeedFunc(func(ctx context.Context, sink func(FileEvent)) error {
		go func() {
			defer close(callbackDone)
			sink(FileEvent{PID: 101, Path: "/owned/canary", Op: OpWrite})
		}()
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, feed, 999, nil) }()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("callback never entered Consider")
	}
	cancel()
	waitAdmissionRun(t, done)
	releaseOnce.Do(func() { close(release) })
	select {
	case <-callbackDone:
	case <-time.After(2 * time.Second):
		t.Fatal("callback did not return")
	}
	if canaries.calls.Load() != 1 || resp.killCount() != 0 || len(app.lines) != 0 || len(incidents.recs) != 0 {
		t.Fatal("cancellation during filtering admitted evidence or response")
	}
	g.scorer.mu.Lock()
	states := len(g.scorer.states)
	g.scorer.mu.Unlock()
	if states != 0 {
		t.Fatal("cancelled callback retained scorer evidence")
	}
}

func TestGuardRetiredFeedCallbackBlockedInConsiderCannotEnterReplacementAttempt(t *testing.T) {
	old := feedRetryBackoff
	feedRetryBackoff = time.Millisecond
	defer func() { feedRetryBackoff = old }()
	g, resp, app, _, incidents, table := newTestGuard("kill", nil, 50, 100)
	table.Add(proctable.Proc{PID: 101, Image: "/owned/harmless-test-process"})
	entered, release, callbackDone := make(chan struct{}), make(chan struct{}), make(chan struct{})
	var enteredOnce, releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	canaries := &admissionCanaries{check: func(string) bool {
		enteredOnce.Do(func() { close(entered) })
		<-release
		return true
	}}
	g.deps.Canaries = canaries
	replacementStarted := make(chan struct{})
	attempts := 0
	feed := admissionFeedFunc(func(ctx context.Context, sink func(FileEvent)) error {
		attempts++
		if attempts == 1 {
			go func() {
				defer close(callbackDone)
				sink(FileEvent{PID: 101, Path: "/owned/canary", Op: OpWrite})
			}()
			select {
			case <-entered:
				return errors.New("feed failed with callback still in filter")
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		close(replacementStarted)
		<-ctx.Done()
		return ctx.Err()
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- g.Run(ctx, feed, 999, nil) }()
	select {
	case <-replacementStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("failed feed callback prevented a bounded retry")
	}
	if ctx.Err() != nil {
		t.Fatal("test must keep parent context live during retry")
	}
	releaseOnce.Do(func() { close(release) })
	select {
	case <-callbackDone:
	case <-time.After(2 * time.Second):
		t.Fatal("retired callback did not return")
	}
	if canaries.calls.Load() != 1 || resp.killCount() != 0 || len(app.lines) != 0 || len(incidents.recs) != 0 {
		t.Fatal("old callback crossed its cancelled attempt into live replacement")
	}
	cancel()
	waitAdmissionRun(t, done)
}
