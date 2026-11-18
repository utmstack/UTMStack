//go:build linux

package watcher_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

type pipelineGate struct {
	scanner *scanner.Scanner
	once    sync.Once
	read    chan string
	release chan struct{}
}

func (g *pipelineGate) ScanFile(path, source string) (string, string, error) {
	v, s, err := g.scanner.ScanFile(path, source)
	g.once.Do(func() { g.read <- v; <-g.release })
	return v, s, err
}

type pipelineSink struct {
	orch *orchestrator.Orchestrator
	seen chan string
}

func (s pipelineSink) Enqueue(path, op string) {
	s.orch.Enqueue(orchestrator.FileEvent{Path: path, Op: op})
	select {
	case s.seen <- path:
	default:
	}
}

// Exercise actual inotify -> common queue -> native engine -> verified
// quarantine. The only pause is after the first real clean scan returns, so a
// rewritten file's kernel notification deterministically overlaps that scan.
func TestLinuxInotifyRewriteAfterSnapshotIsContained(t *testing.T) {
	base := t.TempDir()
	inbox := filepath.Join(base, "inbox")
	db := filepath.Join(base, "signatures")
	for _, dir := range []string{inbox, db} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	marker := []byte("OPENEDR-INERT-REWRITE-QUEUE-CONTROL-20261111")
	if err := os.WriteFile(filepath.Join(db, "control.ndb"), []byte("Lab.Queue.Rewrite:0:*:"+hex.EncodeToString(marker)+"\n"), 0600); err != nil {
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
	target := filepath.Join(inbox, "rewritten.txt")
	if err := os.WriteFile(target, []byte("ordinary initial clean bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	g := &pipelineGate{scanner: scanner.New(cfg, c, sp, q, eng), read: make(chan string, 1), release: make(chan struct{})}
	o := orchestrator.New(g, nil, 1, 1)
	ctx, cancel := context.WithCancel(context.Background())
	o.Start(ctx)
	sink := pipelineSink{o, make(chan string, 32)}
	w := watcher.New([]string{inbox}, sink, nil, nil)
	done := make(chan error, 1)
	go func() { done <- w.RunError(ctx) }()
	var releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(g.release) }) }
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
	case v := <-g.read:
		if v != cache.VerdictClean {
			t.Fatalf("first real scan=%q", v)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("first snapshot missing")
	}
	deadline := time.Now().Add(time.Second)
	for w.Health() != nil {
		if time.Now().After(deadline) {
			t.Fatal(w.Health())
		}
		time.Sleep(time.Millisecond)
	}
	for len(sink.seen) > 0 {
		<-sink.seen
	}
	if err := os.WriteFile(target, marker, 0600); err != nil {
		t.Fatal(err)
	}
	select {
	case path := <-sink.seen:
		if path != target {
			t.Fatal(path)
		}
	case <-time.After(time.Second):
		t.Fatal("kernel rewrite notification missing")
	}
	started := time.Now()
	release()
	deadline = time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("kernel-observed rewrite survived without a follow-up containment")
		}
		time.Sleep(time.Millisecond)
	}
	// Wait for response metadata, which follows the physical move.
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
			t.Fatal("quarantine metadata missing")
		}
		time.Sleep(time.Millisecond)
	}
	sum := sha256.Sum256(marker)
	if rows[0].OriginalPath != target || rows[0].SHA256 != hex.EncodeToString(sum[:]) || rows[0].Detection != "Lab.Queue.Rewrite" {
		t.Fatalf("wrong admitted object: %+v", rows)
	}
	stored, err := os.ReadFile(filepath.Join(qdir, rows[0].QuarantineID+".quarantined"))
	if err != nil || string(stored) != string(marker) {
		t.Fatalf("wrong stored snapshot: %q %v", stored, err)
	}
	b, _ := json.Marshal(map[string]any{"first_verdict": cache.VerdictClean, "kernel_rewrite_observed": true, "contained_sha256": rows[0].SHA256, "latency_ms": time.Since(started).Milliseconds(), "engine_generation": eng.Version()})
	t.Log(string(b))
}
