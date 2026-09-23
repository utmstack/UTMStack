package orchestrator

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

// fakeScanner is the hermetic seam for these tests: it records every path the
// shared pool asks it to scan and returns a per-path (or default) verdict — no
// engine calls.
type fakeScanner struct {
	mu      sync.Mutex
	scanned []string
	verdict map[string]string // path → verdict; absent = clean
	delay   time.Duration     // per-scan delay (to pace the pool)
}

func (f *fakeScanner) ScanFile(path, source string) (string, string, error) {
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	f.mu.Lock()
	f.scanned = append(f.scanned, path)
	f.mu.Unlock()
	v, ok := f.verdict[path]
	if !ok {
		v = cache.VerdictClean
	}
	return v, "test-sig", nil
}

func (f *fakeScanner) seen() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.scanned...)
}

// newTestPool builds a started orchestrator whose scanner is a counting wrapper
// over fake. Workers are torn down (and the ctx cancelled) in cleanup.
func newTestPool(t *testing.T, fake *fakeScanner, ex *Excluder, workers int) (*Orchestrator, *ScanStats) {
	t.Helper()
	stats := &ScanStats{}
	o := New(NewCountingScanner(fake, stats), ex, workers, 4096)
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	t.Cleanup(func() {
		cancel()
		o.Stop()
	})
	return o, stats
}

// mustSpool opens a real spool under dir (hermetic) and closes it in cleanup.
func mustSpool(t *testing.T, dir string) *event.Spool {
	t.Helper()
	p := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(p, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp.Close() })
	return sp
}

// writeFiles creates dir/subtree as needed and writes a file at each name.
func writeFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, n := range names {
		p := filepath.Join(dir, n)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// awaitScanned waits (bounded) until the pool has scanned at least n paths.
func awaitScanned(t *testing.T, fake *fakeScanner, n int, d time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if len(fake.seen()) >= n {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return false
}

// inflightCount is a test hook into the orchestrator's coalescing map: with a
// blocked worker it equals the number of items the walk enqueued so far.
func inflightCount(o *Orchestrator) int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return len(o.inflight)
}

// Test 1: the walk enqueues only non-excluded regular files across multiple
// target dirs. Excluded files, symlinks, and empty dirs are skipped.
func TestScheduledWalkEnqueuesOnlyNonExcludedRegular(t *testing.T) {
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	e := filepath.Join(root, "empty") // dir with no files
	excl := filepath.Join(root, "excl")
	os.MkdirAll(e, 0o755)
	writeFiles(t, a, "keep1.txt", "keep2.txt")
	writeFiles(t, b, "keep3.txt")
	writeFiles(t, excl, "skipme.txt")
	if err := os.Symlink(filepath.Join(a, "keep1.txt"), filepath.Join(b, "link.txt")); err != nil {
		t.Skipf("symlinks unsupported here: %v", err)
	}

	fake := &fakeScanner{}
	o, _ := newTestPool(t, fake, NewExcluder([]string{excl}), 2)
	sp := mustSpool(t, root)
	ss := NewScheduledScanner(ScheduledScanParams{
		Service: context.Background(),
		Orch:    o, Ex: NewExcluder([]string{excl}), Paths: []string{a, b, e},
		Stats: &ScanStats{}, Spool: sp,
	})
	if !ss.Trigger("every:1h") {
		t.Fatal("Trigger = false, want true")
	}
	ss.Wait() // let the walk finish; then drain the pool

	if !awaitScanned(t, fake, 3, 2*time.Second) {
		t.Fatalf("pool scanned %v, want 3 (keep1, keep2, keep3)", fake.seen())
	}
	// A little grace for any (incorrectly enqueued) extra items to surface.
	time.Sleep(50 * time.Millisecond)
	got := map[string]bool{}
	for _, p := range fake.seen() {
		got[p] = true
	}
	if len(got) != 3 {
		t.Fatalf("scanned set = %v, want exactly keep1/keep2/keep3", fake.seen())
	}
	if got[filepath.Join(excl, "skipme.txt")] {
		t.Fatal("excluded file was scanned")
	}
	if got[filepath.Join(b, "link.txt")] {
		t.Fatal("symlink was scanned (must be treated as non-regular)")
	}
}

// Test 2: at most one walk runs at a time. A trigger while running is a no-op;
// after the run finishes, the next trigger proceeds.
func TestScheduledOneAtATime(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	for i := 0; i < 30; i++ {
		writeFiles(t, dir, "s", "f"+strconv.Itoa(i)+".txt")
	}
	fake := &fakeScanner{}
	o, _ := newTestPool(t, fake, NewExcluder(nil), 1)
	sp := mustSpool(t, root)
	ss := NewScheduledScanner(ScheduledScanParams{
		Service: context.Background(),
		Orch:    o, Ex: NewExcluder(nil), Paths: []string{dir},
		Batch: 1, Throttle: 10 * time.Millisecond,
		Stats: &ScanStats{}, Spool: sp,
	})
	if !ss.Trigger("every:1h") {
		t.Fatal("first Trigger = false, want true")
	}
	if ss.Trigger("every:1h") {
		t.Fatal("second Trigger while running = true, want false (no-op)")
	}
	ss.Cancel() // wait for the in-flight walk to end
	if !ss.Trigger("every:1h") {
		t.Fatal("Trigger after completion = false, want true (guard released)")
	}
	ss.Cancel()
}

// Test 3: cancelling mid-run stops the walk promptly and the run returns. The
// worker is blocked so nothing drains: inflight count equals enqueued count and
// must stop growing the moment the walk is cancelled.
func TestScheduledCancelMidRun(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	const n = 400
	for i := 0; i < n; i++ {
		writeFiles(t, dir, "b", "f"+strconv.Itoa(i)+".txt")
	}
	// A single worker blocked forever on the first scan: every enqueued item
	// stays in the queue/inflight, so enqueued count is observable.
	fake := &fakeScanner{}
	release := make(chan struct{})
	blocking := &blockingScanner{fake: fake, release: release}
	stats := &ScanStats{}
	o := New(blocking, NewExcluder(nil), 1, 4096)
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	defer func() {
		close(release)
		cancel()
		o.Stop()
	}()
	_ = stats
	sp := mustSpool(t, root)
	ss := NewScheduledScanner(ScheduledScanParams{
		Service: context.Background(),
		Orch:    o, Ex: NewExcluder(nil), Paths: []string{dir},
		Batch: 4, Throttle: 15 * time.Millisecond,
		Stats: &ScanStats{}, Spool: sp,
	})
	if !ss.Trigger("every:1h") {
		t.Fatal("Trigger = false, want true")
	}
	time.Sleep(200 * time.Millisecond) // let the walk enqueue a handful of batches
	start := time.Now()
	ss.Cancel() // must stop the walk and return — bounded, no hang
	if took := time.Since(start); took > 2*time.Second {
		t.Fatalf("Cancel took %v, want < 2s", took)
	}
	first := inflightCount(o)
	time.Sleep(150 * time.Millisecond) // walk is over; nothing new may arrive
	second := inflightCount(o)
	if second != first {
		t.Fatalf("enqueueing continued after cancel: %d → %d", first, second)
	}
	if second >= n {
		t.Fatalf("walk did not stop early: enqueued %d of %d", second, n)
	}
}

// blockingScanner forwards to a fake after blocking on release, so the pool
// holds every path in the inflight map (observable enqueue count).
type blockingScanner struct {
	fake    *fakeScanner
	release chan struct{}
}

func (b *blockingScanner) ScanFile(path, source string) (string, string, error) {
	<-b.release
	return b.fake.ScanFile(path, source)
}

// Test 4: throttle — 10 files, batch 2, 20ms pause → at least 5 pauses, so the
// walk (and thus Cancel, which waits for it) takes >= 100ms.
func TestScheduledThrottlePace(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	const n = 10
	for i := 0; i < n; i++ {
		writeFiles(t, dir, "f"+strconv.Itoa(i)+".txt")
	}
	fake := &fakeScanner{}
	o, _ := newTestPool(t, fake, NewExcluder(nil), 2)
	sp := mustSpool(t, root)
	ss := NewScheduledScanner(ScheduledScanParams{
		Service: context.Background(),
		Orch:    o, Ex: NewExcluder(nil), Paths: []string{dir},
		Batch: 2, Throttle: 20 * time.Millisecond,
		Stats: &ScanStats{}, Spool: sp,
	})
	start := time.Now()
	if !ss.Trigger("every:1h") {
		t.Fatal("Trigger = false, want true")
	}
	ss.Wait() // let the walk run to completion
	elapsed := time.Since(start)
	min := time.Duration(n/2) * 20 * time.Millisecond
	if elapsed < min {
		t.Fatalf("walk elapsed %v, want >= %v (%d files, batch 2, 20ms pause)", elapsed, min, n)
	}
}

// Test 5: the runner appends a scan_start and a scan_complete event to a real
// spool. The completion event carries files_scanned, elapsed_ms, and — with the
// counting wrapper in the path — the run's detections.
func TestScheduledEmitsStartAndCompleteEvents(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "d")
	const n = 8
	for i := 0; i < n; i++ {
		writeFiles(t, dir, "f"+strconv.Itoa(i)+".txt")
	}
	bad := filepath.Join(dir, "f1.txt")
	fake := &fakeScanner{delay: 50 * time.Millisecond, verdict: map[string]string{bad: cache.VerdictMalicious}}
	o, stats := newTestPool(t, fake, NewExcluder(nil), 2)
	sp := mustSpool(t, root)
	// Throttle the walk well past the pool's drain time (8 files × 50ms on 2
	// workers ≈ 200ms) so the completion snapshot already includes the
	// malicious verdict.
	ss := NewScheduledScanner(ScheduledScanParams{
		Service: context.Background(),
		Orch:    o, Ex: NewExcluder(nil), Paths: []string{dir},
		Batch: n, Throttle: 500 * time.Millisecond,
		Stats: stats, Spool: sp,
	})
	if !ss.Trigger("every:1h") {
		t.Fatal("Trigger = false, want true")
	}
	ss.Wait()                         // let the walk finish so the completion event carries real stats
	time.Sleep(50 * time.Millisecond) // grace for the last scan's stat write

	raw, err := os.ReadFile(sp.Path())
	if err != nil {
		t.Fatal(err)
	}
	var startEv, endEv map[string]any
	for _, line := range splitLines(string(raw)) {
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			t.Fatalf("bad spool line %q: %v", line, err)
		}
		switch m["action"] {
		case event.ActionScanStart:
			startEv = m
		case event.ActionScanComplete:
			endEv = m
		}
	}
	if startEv == nil {
		t.Fatal("no scan_start event in spool")
	}
	if startEv["source"] != event.SourceScheduled {
		t.Fatalf("start source = %v, want %s", startEv["source"], event.SourceScheduled)
	}
	if endEv == nil {
		t.Fatal("no scan_complete event in spool")
	}
	if got := endEv["files_scanned"]; got != float64(n) {
		t.Fatalf("files_scanned = %v, want %d", got, n)
	}
	if _, ok := endEv["elapsed_ms"]; !ok {
		t.Fatal("scan_complete missing elapsed_ms")
	}
	if got := endEv["detections"]; got != float64(1) {
		t.Fatalf("detections = %v, want 1 (counting wrapper in the path)", got)
	}
}

// Test 6: parseSchedule accepts only "every:<duration>".
func TestParseScheduleForms(t *testing.T) {
	d, err := ParseSchedule("every:24h")
	if err != nil || d != 24*time.Hour {
		t.Fatalf("ParseSchedule(every:24h) = %v, %v", d, err)
	}
	if _, err := ParseSchedule("24h"); err == nil {
		t.Fatal("missing prefix must be rejected")
	}
	if _, err := ParseSchedule("daily"); err == nil {
		t.Fatal("unknown form must be rejected")
	}
	if _, err := ParseSchedule("every:0s"); err == nil {
		t.Fatal("zero/negative duration must be rejected")
	}
	if _, err := ParseSchedule("every:banana"); err == nil {
		t.Fatal("unparseable duration must be rejected")
	}
}

// Test 7: countingScanner increments the shared stats and delegates the
// verdict/signature to the inner scanner.
func TestCountingScannerCountsAndDelegates(t *testing.T) {
	clean := &fakeScanner{}
	stats := &ScanStats{}
	w := NewCountingScanner(clean, stats).(*countingScanner)
	v, sig, err := w.ScanFile("/a", "src")
	if err != nil || v != cache.VerdictClean || sig != "test-sig" {
		t.Fatalf("delegated = (%q, %q, %v), want (%s, test-sig, nil)", v, sig, err, cache.VerdictClean)
	}
	if tot, mal := stats.Snapshot(); tot != 1 || mal != 0 {
		t.Fatalf("stats after clean scan = (%d, %d), want (1, 0)", tot, mal)
	}

	bad := &fakeScanner{verdict: map[string]string{"/b": cache.VerdictMalicious}}
	stats2 := &ScanStats{}
	w2 := NewCountingScanner(bad, stats2).(*countingScanner)
	if v, _, _ = w2.ScanFile("/b", "src"); v != cache.VerdictMalicious {
		t.Fatalf("verdict = %q, want malicious (delegated)", v)
	}
	if v, _, _ = w2.ScanFile("/c", "src"); v != cache.VerdictClean {
		t.Fatalf("verdict = %q, want clean", v)
	}
	if tot, mal := stats2.Snapshot(); tot != 2 || mal != 1 {
		t.Fatalf("stats = (%d, %d), want (2, 1)", tot, mal)
	}
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
