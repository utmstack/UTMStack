package orchestrator

import (
	"context"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

type FileEvent struct {
	Path string
	Op   string
}

type FileScanner interface {
	ScanFile(path, source string) (string, string, error)
}

type Orchestrator struct {
	sc      FileScanner
	ex      *Excluder
	workers int
	queue   chan FileEvent
	wg      sync.WaitGroup

	// onMalicious, if set, is invoked with the path+signature of a file the
	// worker just found malicious. The service wires it to terminate any running
	// process launched from that image, so a malicious executable that is dropped
	// AND immediately run is both quarantined (file) and killed (process) — the
	// file-watcher scan otherwise only removes the on-disk image while the process
	// keeps running (and can outrace the process guard, which loses its scan
	// target once the image is quarantined).
	onMalicious func(path, signature string)

	// inflight coalesces duplicate events for the same path. A single file
	// arrival produces several USN records (create/extend/close), and without
	// this the worker pool would scan the same file concurrently and emit
	// duplicate (and racily contradictory) events. A path is marked inflight
	// from Enqueue until its scan completes; duplicates in that window are dropped.
	mu       sync.Mutex
	inflight map[string]bool
}

func New(sc FileScanner, ex *Excluder, workers, queueSize int) *Orchestrator {
	if workers < 1 {
		workers = 1
	}
	return &Orchestrator{
		sc:       sc,
		ex:       ex,
		workers:  workers,
		queue:    make(chan FileEvent, queueSize),
		inflight: make(map[string]bool),
	}
}

// SetOnMalicious registers a callback invoked with the path+signature of a file
// found malicious by a worker (used to kill processes running that image).
func (o *Orchestrator) SetOnMalicious(fn func(path, signature string)) { o.onMalicious = fn }

func (o *Orchestrator) Start(ctx context.Context) {
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(ctx)
	}
}

func (o *Orchestrator) worker(ctx context.Context) {
	defer o.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-o.queue:
			if o.ex == nil || !o.ex.Excluded(ev.Path) {
				// Errors (file vanished, engine down) are non-fatal; the file stays
				// unknown and will be revisited on the next event/sweep.
				verdict, sig, _ := o.sc.ScanFile(ev.Path, event.SourceFileWatcher)
				if verdict == cache.VerdictMalicious && o.onMalicious != nil {
					o.onMalicious(ev.Path, sig)
				}
			}
			o.done(ev.Path)
		}
	}
}

// Enqueue returns false when the queue is full (caller logs the drop). Duplicate
// events for a path already queued/in-flight are coalesced (dropped, reported as
// accepted) so one file arrival yields a single scan and a single event.
func (o *Orchestrator) Enqueue(ev FileEvent) bool {
	o.mu.Lock()
	if o.inflight[ev.Path] {
		o.mu.Unlock()
		return true // already queued or being scanned; coalesce
	}
	o.inflight[ev.Path] = true
	o.mu.Unlock()

	select {
	case o.queue <- ev:
		return true
	default:
		o.done(ev.Path) // queue full; release so a later event can retry
		return false
	}
}

func (o *Orchestrator) done(path string) {
	o.mu.Lock()
	delete(o.inflight, path)
	o.mu.Unlock()
}

func (o *Orchestrator) Stop() { o.wg.Wait() }
