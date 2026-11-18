//go:build linux

package procwatch

import (
	"context"
	"errors"
	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"github.com/utmstack/UTMStack/shared/logger"
	"sync"
	"time"
)

type Watcher struct {
	h             Handler
	mu            sync.Mutex
	health        error
	reader        linuxproc.Reader
	interval      time.Duration
	requestedMode string
	mode          string
	coverage      error
	misses        uint64
}

func New(h Handler) *Watcher {
	return NewWithMode(h, "auto")
}
func NewWithMode(h Handler, mode string) *Watcher {
	return &Watcher{h: h, requestedMode: mode, health: errors.New("process monitor not started"), interval: 250 * time.Millisecond}
}
func (w *Watcher) Health() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return errors.Join(w.health, w.coverage)
}
func (w *Watcher) setHealth(err error) { w.mu.Lock(); w.health = err; w.mu.Unlock() }
func (w *Watcher) setMode(mode string) { w.mu.Lock(); w.mode = mode; w.mu.Unlock() }
func (w *Watcher) Mode() string        { w.mu.Lock(); defer w.mu.Unlock(); return w.mode }
func (w *Watcher) recordMiss(err error) {
	w.mu.Lock()
	w.misses++
	w.coverage = errors.New("process connector metadata unavailable for an observed event; coverage incomplete")
	w.mu.Unlock()
}
func (w *Watcher) Run(ctx context.Context) {
	for {
		err := w.RunError(ctx)
		if ctx.Err() != nil {
			return
		}
		logger.Error("UTMStack EDR process monitor unavailable: %v", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}
func (w *Watcher) RunError(ctx context.Context) (err error) {
	defer func() { w.setHealth(err) }()
	if w.h == nil {
		return errors.New("process monitor requires a handler")
	}
	switch w.requestedMode {
	case "auto", "connector", "":
		err = w.runConnector(ctx)
		if ctx.Err() != nil || w.requestedMode == "connector" {
			return err
		}
		w.mu.Lock()
		w.coverage = errors.New("process connector unavailable; using procfs polling with short-process gaps: " + err.Error())
		w.mu.Unlock()
	case "poll":
	default:
		return errors.New("invalid Linux process monitor mode")
	}
	w.setMode("poll")
	return w.runPolling(ctx)
}
func (w *Watcher) runPolling(ctx context.Context) error {
	previous, err := w.reader.Snapshot(ctx, true)
	if err != nil {
		return err
	}
	seen, _ := processChanges(nil, previous, true)
	for _, p := range previous {
		if p.PID > 1 {
			w.h.OnProcStart(ProcStart{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline, OSIdentity: p.StartTicks, IdentityRequired: true, EventKind: "inventory"})
		}
	}
	w.setHealth(nil)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			current, err := w.reader.Snapshot(ctx, true)
			if err != nil {
				return err
			}
			next, changes := processChanges(seen, current, false)
			if len(changes) > 1024 {
				return errors.New("process monitor launch burst exceeds 1024 observations")
			}
			if exits, ok := w.h.(interface{ OnProcExit(int, int64) }); ok {
				for pid, old := range seen {
					if now, exists := next[pid]; !exists || now.identity != old.identity {
						exits.OnProcExit(pid, old.identity)
					}
				}
			}
			for _, p := range changes {
				if err := ctx.Err(); err != nil {
					return err
				}
				p.EventKind = "snapshot"
				w.h.OnProcStart(p)
			}
			seen = next
		}
	}
}
