//go:build linux

package watcher_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
	"github.com/utmstack/UTMStack/agent/edr/quarantine"
	"github.com/utmstack/UTMStack/agent/edr/scanner"
	"github.com/utmstack/UTMStack/agent/edr/watcher"
)

type pressurePipelineSink struct {
	orch     *orchestrator.Orchestrator
	mu       sync.Mutex
	rejected map[string]int
}

func (s *pressurePipelineSink) Enqueue(path, op string) { s.TryEnqueue(path, op) }
func (s *pressurePipelineSink) TryEnqueue(path, op string) bool {
	accepted := s.orch.Enqueue(orchestrator.FileEvent{Path: path, Op: op})
	if !accepted {
		s.mu.Lock()
		s.rejected[path]++
		s.mu.Unlock()
	}
	return accepted
}
func (s *pressurePipelineSink) rejectionCount(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rejected[path]
}

// A full real queue must trigger reconciliation of persistent bytes. The first
// actual scan is held after its clean snapshot so notification admission is
// deterministically exhausted; no fake engine verdict or kernel event is used.
func TestLinuxInotifyQueuePressureRecoversPersistentMaliciousFile(t *testing.T) {
	base := t.TempDir()
	inbox, db := filepath.Join(base, "inbox"), filepath.Join(base, "signatures")
	for _, dir := range []string{inbox, db} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	marker := []byte("OPENEDR-INERT-QUEUE-RECOVERY-CONTROL-20261112")
	if err := os.WriteFile(filepath.Join(db, "control.ndb"), []byte("Lab.Queue.Recovery:0:*:"+hex.EncodeToString(marker)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir, cfg.ScanConcurrency = db, 1
	eng := engine.New(cfg)
	t.Cleanup(func() { eng.Stop() })
	if err := eng.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(base, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	sp, err := event.OpenSpool(filepath.Join(base, "events.ndjson"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sp.Close() })
	qdir := filepath.Join(base, "quarantine")
	q, err := quarantine.New(qdir, c)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(inbox, "initial.txt"), []byte("ordinary initial bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	g := &pipelineGate{scanner: scanner.New(cfg, c, sp, q, eng), read: make(chan string, 1), release: make(chan struct{})}
	o := orchestrator.New(g, nil, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	sink := &pressurePipelineSink{orch: o, rejected: make(map[string]int)}
	w := watcher.New([]string{inbox}, sink, nil, nil)
	done := make(chan struct{})
	go func() { defer close(done); w.Run(ctx) }()
	var once sync.Once
	release := func() { once.Do(func() { close(g.release) }) }
	t.Cleanup(func() {
		release()
		cancel()
		o.Stop()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Error("watcher did not stop")
		}
	})
	select {
	case verdict := <-g.read:
		if verdict != cache.VerdictClean {
			t.Fatalf("first scan=%q", verdict)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("first real snapshot missing")
	}
	deadline := time.Now().Add(3 * time.Second)
	for w.Health() != nil {
		if time.Now().After(deadline) {
			t.Fatal("initial watcher not active")
		}
		time.Sleep(time.Millisecond)
	}
	for i := 0; i < 20; i++ {
		if err := os.WriteFile(filepath.Join(inbox, fmt.Sprintf("burst-%02d.txt", i)), []byte(fmt.Sprintf("owned benign bytes %d", i)), 0600); err != nil {
			t.Fatal(err)
		}
	}
	target := filepath.Join(inbox, "persistent-marked.txt")
	if err := os.WriteFile(target, marker, 0600); err != nil {
		t.Fatal(err)
	}
	// The old implementation consumes and rejects both target notifications.
	// A recovering implementation stops the stream on the first rejection and
	// exposes its coverage gap before retrying the full directory.
	gapReported := false
	deadline = time.Now().Add(4 * time.Second)
	for {
		if o.Snapshot().Rejected > 0 && w.Health() != nil {
			gapReported = true
			break
		}
		if sink.rejectionCount(target) >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("pressure not observed: %+v", o.Snapshot())
		}
		time.Sleep(time.Millisecond)
	}
	started := time.Now()
	release()
	deadline = time.Now().Add(12 * time.Second)
	var rows []cache.QuarantineRecord
	for {
		rows, err = c.ListQuarantine()
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) == 1 {
			break
		}
		if time.Now().After(deadline) {
			remaining, readErr := os.ReadFile(target)
			sum := sha256.Sum256(remaining)
			t.Fatalf("persistent malicious bytes survived queue pressure: rejected_target=%d gap_reported=%v queue=%+v remaining_sha256=%x marker_matches=%v read_error=%v", sink.rejectionCount(target), gapReported, o.Snapshot(), sum, string(remaining) == string(marker), readErr)
		}
		time.Sleep(5 * time.Millisecond)
	}
	if !gapReported {
		t.Fatal("queue rejection did not expose a reconciliation gap")
	}
	sum := sha256.Sum256(marker)
	if rows[0].OriginalPath != target || rows[0].SHA256 != hex.EncodeToString(sum[:]) || rows[0].Detection != "Lab.Queue.Recovery" {
		t.Fatalf("wrong admitted object: %+v", rows)
	}
	stored, err := os.ReadFile(filepath.Join(qdir, rows[0].QuarantineID+".quarantined"))
	if err != nil || string(stored) != string(marker) {
		t.Fatalf("wrong retained snapshot: %q %v", stored, err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("source was not contained: %v", err)
	}
	result, _ := json.Marshal(map[string]any{"gap_reported": gapReported, "contained_sha256": rows[0].SHA256, "latency_ms": time.Since(started).Milliseconds(), "queue": o.Snapshot(), "engine_generation": eng.Version()})
	t.Log(string(result))
}
