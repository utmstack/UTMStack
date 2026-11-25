package eventstore

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/go-sdk/v2/store"
)

type slowFields struct {
	store.Store
	calls atomic.Int32
	delay time.Duration
	out   []store.Field
}

func (f *slowFields) DescribeFields(ctx context.Context, _ store.Scope) ([]store.Field, error) {
	f.calls.Add(1)

	select {
	case <-time.After(f.delay):
		return f.out, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func newStore(f *slowFields) *Store { return &Store{Store: f} }

var scope = store.Scope{Dataset: DatasetLogs, Tenant: "t1"}

func TestFieldsAreComputedOnceForEveryoneAsking(t *testing.T) {
	f := &slowFields{delay: 80 * time.Millisecond, out: []store.Field{{Name: "origin.user"}}}
	s := newStore(f)

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			if got, err := s.DescribeFields(context.Background(), scope); err != nil || len(got) != 1 {
				t.Errorf("%v %v", got, err)
			}
		}()
	}

	wg.Wait()

	if n := f.calls.Load(); n != 1 {
		t.Fatalf("the table was read %d times, want 1", n)
	}

	if _, err := s.DescribeFields(context.Background(), scope); err != nil || f.calls.Load() != 1 {
		t.Fatalf("a second look within the TTL must not read again: %d", f.calls.Load())
	}
}

func TestTenantsHaveTheirOwnCopy(t *testing.T) {
	f := &slowFields{out: []store.Field{{Name: "x"}}}
	s := newStore(f)

	_, _ = s.DescribeFields(context.Background(), scope)
	_, _ = s.DescribeFields(context.Background(), store.Scope{Dataset: DatasetLogs, Tenant: "t2"})

	if f.calls.Load() != 2 {
		t.Fatalf("calls = %d", f.calls.Load())
	}
}

// The first caller gives up after a moment; the work goes on, so the next one is served at once.
func TestACallerThatStopsWaitingDoesNotStopTheWork(t *testing.T) {
	f := &slowFields{delay: 150 * time.Millisecond, out: []store.Field{{Name: "x"}}}
	s := newStore(f)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	if _, err := s.DescribeFields(ctx, scope); err == nil {
		t.Fatal("the impatient caller should have given up")
	}

	time.Sleep(250 * time.Millisecond)

	start := time.Now()
	got, err := s.DescribeFields(context.Background(), scope)

	if err != nil || len(got) != 1 || time.Since(start) > 50*time.Millisecond || f.calls.Load() != 1 {
		t.Fatalf("the finished work must be waiting in the cache: %v %v %s calls=%d", got, err, time.Since(start), f.calls.Load())
	}
}

func TestAStaleCopyIsServedWhileARefreshRuns(t *testing.T) {
	f := &slowFields{delay: 120 * time.Millisecond, out: []store.Field{{Name: "old"}}}
	s := newStore(f)

	_, _ = s.DescribeFields(context.Background(), scope)

	s.fields.mu.Lock()
	e := s.fields.items["logs|t1|"]
	e.at = time.Now().Add(-fieldsTTL - time.Minute)
	s.fields.items["logs|t1|"] = e
	s.fields.mu.Unlock()

	f.out = []store.Field{{Name: "new"}}

	start := time.Now()
	got, _ := s.DescribeFields(context.Background(), scope)

	if got[0].Name != "old" || time.Since(start) > 50*time.Millisecond {
		t.Fatalf("expected the old copy at once, got %v after %s", got, time.Since(start))
	}

	time.Sleep(250 * time.Millisecond)

	if got, _ := s.DescribeFields(context.Background(), scope); got[0].Name != "new" {
		t.Fatalf("the refresh should have landed: %v", got)
	}
}
