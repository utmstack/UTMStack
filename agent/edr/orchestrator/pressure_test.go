package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stepScanner struct {
	seen chan FileEvent
	step chan struct{}
}

func (s *stepScanner) ScanFile(path, source string) (string, string, error) {
	s.seen <- FileEvent{Path: path, Source: source}
	<-s.step
	if path == "error" {
		return "", "", errors.New("analysis unavailable")
	}
	return "clean", "", nil
}
func nextScan(t *testing.T, s *stepScanner, path, source string) {
	t.Helper()
	select {
	case ev := <-s.seen:
		if ev.Path != path || ev.Source != source {
			t.Fatalf("scan=%+v want %s/%s", ev, path, source)
		}
	case <-time.After(time.Second):
		t.Fatalf("no scan for %s", path)
	}
}
func drained(t *testing.T, o *Orchestrator, n uint64) Stats {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for {
		s := o.Snapshot()
		if s.Completed == n && s.Active == 0 && s.Queued == 0 {
			return s
		}
		if time.Now().After(deadline) {
			t.Fatalf("queue not drained: %+v", s)
		}
		time.Sleep(time.Millisecond)
	}
}
func TestQueuedDuplicatesShareFutureSnapshot(t *testing.T) {
	sc := &stepScanner{make(chan FileEvent, 8), make(chan struct{}, 8)}
	o := New(sc, nil, 2, 1)
	for i := 0; i < 8; i++ {
		if !o.Enqueue(FileEvent{Path: "one", Source: "last"}) {
			t.Fatal("duplicate rejected")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { cancel(); o.Stop() }()
	o.Start(ctx)
	nextScan(t, sc, "one", "last")
	sc.step <- struct{}{}
	s := drained(t, o, 1)
	if s.Enqueued != 1 || s.Coalesced != 7 || s.Requeued != 0 || s.Rejected != 0 {
		t.Fatal(s)
	}
}
func TestDirtyReturnIsBoundedAndFollowsQueuedWork(t *testing.T) {
	sc := &stepScanner{make(chan FileEvent, 8), make(chan struct{}, 8)}
	o := New(sc, nil, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer func() { close(sc.step); cancel(); o.Stop() }()
	o.Start(ctx)
	o.Enqueue(FileEvent{Path: "hot", Source: "initial"})
	nextScan(t, sc, "hot", "initial")
	if !o.Enqueue(FileEvent{Path: "error", Source: "queued"}) {
		t.Fatal("queue not admitted")
	}
	for i := 0; i < 100; i++ {
		if !o.Enqueue(FileEvent{Path: "hot", Source: "latest"}) {
			t.Fatal("dirty path rejected")
		}
	}
	if o.Enqueue(FileEvent{Path: "overflow"}) {
		t.Fatal("full queue admitted another distinct path")
	}
	s := o.Snapshot()
	if s.Active != 1 || s.Queued != 1 || s.Capacity != 2 || s.Rejected != 1 {
		t.Fatal(s)
	}
	sc.step <- struct{}{}
	nextScan(t, sc, "error", "queued")
	sc.step <- struct{}{}
	nextScan(t, sc, "hot", "latest")
	// A later write during the follow-up needs another future snapshot.
	o.Enqueue(FileEvent{Path: "hot", Source: "again"})
	sc.step <- struct{}{}
	nextScan(t, sc, "hot", "again")
	sc.step <- struct{}{}
	s = drained(t, o, 4)
	if s.Enqueued != 2 || s.Coalesced != 101 || s.Requeued != 2 || s.Started != 4 || s.ScanErrors != 1 {
		t.Fatal(s)
	}
}
func TestCancellationRejectsArrivalsAndWaitsForActiveScan(t *testing.T) {
	sc := &stepScanner{make(chan FileEvent, 8), make(chan struct{})}
	o := New(sc, nil, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	o.Enqueue(FileEvent{Path: "active", Source: "first"})
	nextScan(t, sc, "active", "first")
	o.Enqueue(FileEvent{Path: "active", Source: "dirty"})
	o.Enqueue(FileEvent{Path: "waiting"})
	cancel()
	if o.Enqueue(FileEvent{Path: "late"}) {
		t.Fatal("canceled queue accepted work")
	}
	done := make(chan struct{})
	go func() { o.Stop(); close(done) }()
	select {
	case <-done:
		close(sc.step)
		t.Fatal("active scan outlived Stop")
	case <-time.After(20 * time.Millisecond):
	}
	close(sc.step)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop did not finish")
	}
	s := o.Snapshot()
	if s.Started != 1 || s.Completed != 1 || s.Requeued != 0 || s.Stopped != 1 || s.Abandoned != 2 || s.Active != 0 || s.Queued != 0 {
		t.Fatal(s)
	}
}
