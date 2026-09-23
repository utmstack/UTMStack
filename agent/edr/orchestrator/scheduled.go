package orchestrator

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/shared/logger"
)

// ScanStats holds shared scan counters. The service wraps the real scanner in
// countingScanner so every scan (real-time OR scheduled) bumps these; the
// scheduled runner snapshots them at window start/end to report its own run.
type ScanStats struct {
	total     atomic.Int64 // every file the wrapped scanner looked at
	malicious atomic.Int64 // files the wrapped scanner returned malicious for
}

// Snapshot returns (total, malicious) as of now.
func (s *ScanStats) Snapshot() (int64, int64) { return s.total.Load(), s.malicious.Load() }

// countingScanner is the ONLY change to how the orchestrator's scanner is
// constructed in startPipeline: it delegates to the real *scanner.Scanner and
// increments ScanStats on each call, so scheduled runs can report detections.
// The inner scanner keeps full behavior (hash/cache/engine/quarantine).
type countingScanner struct {
	inner FileScanner
	stats *ScanStats
}

func (c *countingScanner) ScanFile(path, source string) (string, string, error) {
	verdict, sig, err := c.inner.ScanFile(path, source)
	c.stats.total.Add(1)
	if verdict == cache.VerdictMalicious {
		c.stats.malicious.Add(1)
	}
	return verdict, sig, err
}

// NewCountingScanner returns a FileScanner that delegates to inner and records
// each verdict in stats. The service uses it to construct the orchestrator so
// both the real-time watcher and scheduled scans share one counting path.
func NewCountingScanner(inner FileScanner, stats *ScanStats) FileScanner {
	return &countingScanner{inner: inner, stats: stats}
}

// ParseSchedule accepts only the "every:<duration>" form (e.g. "every:24h").
// Any other form is a clear error so a misconfigured schedule fails loudly at
// trigger time rather than silently skipping runs.
func ParseSchedule(s string) (time.Duration, error) {
	const prefix = "every:"
	if len(s) < len(prefix) || s[:len(prefix)] != prefix {
		return 0, fmt.Errorf("schedule must be \"every:<duration>\", got %q", s)
	}
	d, err := time.ParseDuration(s[len(prefix):])
	if err != nil {
		return 0, fmt.Errorf("invalid schedule duration %q: %v", s[len(prefix):], err)
	}
	if d <= 0 {
		return 0, fmt.Errorf("schedule duration must be positive, got %s", d)
	}
	return d, nil
}

// ScheduledScanParams carries the runtime dependencies for NewScheduledScanner.
// Service is the process-wide context: the walk is derived from it so a service
// shutdown cancels any in-flight walk even before Cancel() is invoked.
type ScheduledScanParams struct {
	Service  context.Context // the program's ctx (shutdown signal)
	Orch     *Orchestrator   // shared worker pool (must already be Started)
	Ex       *Excluder       // shared exclusion matcher (nil = no exclusions)
	Paths    []string        // roots to walk; the service resolves empty → fixed volumes
	Batch    int             // files per batch before a Throttle pause (<=0 → 1)
	Throttle time.Duration   // pause between batches (<=0 → none)
	Stats    *ScanStats      // shared counters (from the counting wrapper)
	Spool    *event.Spool    // where scan_start/scan_complete events are appended
}

// ScheduledScanner runs a periodic full-disk scan by walking target roots and
// enqueuing non-excluded regular files to the SHARED orchestrator pool. It does
// NOT create a second scan path. A single run is guarded by `running`; Cancel
// stops the walk only (already-enqueued items drain through the shared pool).
type ScheduledScanner struct {
	svcCtx   context.Context
	orch     *Orchestrator
	ex       *Excluder
	paths    []string
	batch    int
	throttle time.Duration
	stats    *ScanStats
	spool    *event.Spool

	mu      sync.Mutex
	running bool
	cancel  context.CancelFunc // cancels the in-flight walk's ctx
	wg      sync.WaitGroup     // one in-flight walk, if any
}

func NewScheduledScanner(p ScheduledScanParams) *ScheduledScanner {
	if p.Batch < 1 {
		p.Batch = 1
	}
	if p.Service == nil {
		p.Service = context.Background()
	}
	return &ScheduledScanner{
		svcCtx: p.Service, orch: p.Orch, ex: p.Ex, paths: p.Paths, batch: p.Batch,
		throttle: p.Throttle, stats: p.Stats, spool: p.Spool,
	}
}

// Cancel stops the in-progress walk (if any) and blocks until it returns. It
// never stops the shared worker pool — the real-time watcher still needs it —
// so already-enqueued files drain normally. Safe to call when no run is active.
func (s *ScheduledScanner) Cancel() {
	s.mu.Lock()
	c := s.cancel
	s.mu.Unlock()
	if c != nil {
		c()
	}
	s.wg.Wait()
}

// Wait blocks until the in-progress walk (if any) finishes, WITHOUT cancelling
// it — used to let a run complete and its completion event be written. Cancel
// (stop + wait) is for aborting; Wait is for a graceful join.
func (s *ScheduledScanner) Wait() { s.wg.Wait() }

// Trigger starts one scheduled scan walk (non-blocking: the walk runs in its
// own goroutine) if none is active, and returns true. If a run is already in
// progress, or the schedule string is invalid, it logs and returns false. The
// period is validated here but the cadence is owned by the caller (the service
// loop). The walk derives from the service context, so a service shutdown
// (p.cancel) also stops it; Cancel() adds an explicit stop.
func (s *ScheduledScanner) Trigger(schedule string) bool {
	if _, err := ParseSchedule(schedule); err != nil {
		logger.Info("UTMStack EDR: scheduled scan: %v", err)
		return false
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		logger.Info("UTMStack EDR: scheduled scan already running; skipping this trigger")
		return false
	}
	s.running = true
	walkCtx, cancel := context.WithCancel(s.svcCtx)
	s.cancel = cancel
	// ScanStats uses atomics, so snapshotting needs no lock.
	baseTot, baseMal := s.stats.Snapshot()
	s.wg.Add(1)
	s.mu.Unlock()

	_ = s.spool.Append(s.startEvent())

	go func() {
		defer s.wg.Done()
		started := time.Now()
		enqueued := s.walk(walkCtx)
		elapsedMs := time.Since(started).Milliseconds()
		tot, mal := s.stats.Snapshot()
		detections := int(mal - baseMal)
		logger.Debug(100, "UTMStack EDR: scheduled scan: enqueued %d files, scanned %d, detections %d, elapsed %d ms",
			enqueued, tot-baseTot, detections, elapsedMs)
		_ = s.spool.Append(s.completeEvent(enqueued, detections, elapsedMs))
		s.mu.Lock()
		s.running = false
		s.cancel = nil
		s.mu.Unlock()
	}()
	return true
}

func (s *ScheduledScanner) startEvent() string {
	ev := event.Event{Source: event.SourceScheduled, Action: event.ActionScanStart}
	js, _ := ev.ToJSON()
	return js
}

func (s *ScheduledScanner) completeEvent(filesScanned, detections int, elapsedMs int64) string {
	ev := event.Event{
		Source:       event.SourceScheduled,
		Action:       event.ActionScanComplete,
		FilesScanned: filesScanned,
		Detections:   detections,
		ElapsedMs:    elapsedMs,
	}
	js, _ := ev.ToJSON()
	return js
}

// walk traverses each target root and enqueues regular, non-excluded files to
// the shared orchestrator pool at the configured throttle. Returns the count
// of files enqueued. A cancelled ctx stops the walk promptly (via SkipAll) but
// leaves already-enqueued items to drain through the pool.
func (s *ScheduledScanner) walk(ctx context.Context) int {
	enqueued := 0
	for _, root := range s.paths {
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, werr error) error {
			if werr != nil {
				return nil // unreadable entry: skip, don't abort the walk
			}
			select {
			case <-ctx.Done():
				return filepath.SkipAll
			default:
			}
			if !d.Type().IsRegular() {
				return nil // dirs, symlinks, devices: not regular files
			}
			if s.ex != nil && s.ex.Excluded(path) {
				return nil
			}
			if !s.orch.Enqueue(FileEvent{Path: path, Op: "scan"}) {
				logger.Debug(100, "UTMStack EDR: scheduled scan queue full; dropped %s", path)
				return nil
			}
			enqueued++
			if s.throttle > 0 && enqueued%s.batch == 0 {
				select {
				case <-ctx.Done():
					return filepath.SkipAll
				case <-time.After(s.throttle):
				}
			}
			return nil
		})
		if err != nil && err != filepath.SkipAll {
			logger.Error("UTMStack EDR: scheduled scan walk %s: %v", root, err)
		}
		if ctx.Err() != nil {
			break
		}
	}
	return enqueued
}
