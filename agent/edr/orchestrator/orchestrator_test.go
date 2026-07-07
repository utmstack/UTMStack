package orchestrator

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recScanner struct {
	mu   sync.Mutex
	seen []string
}

func (r *recScanner) ScanFile(path, source string) (string, string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen = append(r.seen, path)
	return "clean", "", nil
}

func TestOrchestratorScansEnqueuedNonExcluded(t *testing.T) {
	rs := &recScanner{}
	o := New(rs, NewExcluder([]string{`*.log`}), 2, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	o.Start(ctx)

	o.Enqueue(FileEvent{Path: `C:\a.exe`})
	o.Enqueue(FileEvent{Path: `C:\b.log`}) // excluded → never scanned
	o.Enqueue(FileEvent{Path: `C:\c.dll`})

	// wait for workers to drain
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rs.mu.Lock()
		n := len(rs.seen)
		rs.mu.Unlock()
		if n >= 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if len(rs.seen) != 2 {
		t.Fatalf("scanned %v, want a.exe and c.dll only", rs.seen)
	}
}

// gateScanner blocks in ScanFile until released, so a test can hold a path
// "in flight" while it enqueues duplicates.
type gateScanner struct {
	mu      sync.Mutex
	seen    []string
	release chan struct{}
}

func (g *gateScanner) ScanFile(path, source string) (string, string, error) {
	g.mu.Lock()
	g.seen = append(g.seen, path)
	g.mu.Unlock()
	<-g.release
	return "clean", "", nil
}

// A single file arrival yields several USN records for the same path. The
// orchestrator must coalesce them into one scan/event, not one per record.
func TestOrchestratorCoalescesDuplicatePaths(t *testing.T) {
	gs := &gateScanner{release: make(chan struct{})}
	o := New(gs, nil, 4, 16)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	o.Start(ctx)

	// Enqueue the same path many times while the first scan is blocked in flight.
	for i := 0; i < 8; i++ {
		o.Enqueue(FileEvent{Path: `C:\dup.exe`})
	}
	time.Sleep(100 * time.Millisecond) // let a worker pick it up and block

	gs.mu.Lock()
	n := len(gs.seen)
	gs.mu.Unlock()
	if n != 1 {
		close(gs.release)
		t.Fatalf("duplicate path scanned %d times, want exactly 1", n)
	}
	close(gs.release)
}
