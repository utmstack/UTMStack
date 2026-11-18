package orchestrator

import (
	"context"
	"sync"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

type FileEvent struct {
	Path   string
	Op     string
	Source string
}

type FileScanner interface {
	ScanFile(path, source string) (string, string, error)
}

// Stats counts queue work, not unique objects or kernel observations. Rejected
// events require a later event or reconciliation; a drained queue cannot undo
// that coverage loss. Counters last for this service instance.
type Stats struct {
	Enqueued   uint64 `json:"enqueued"`
	Coalesced  uint64 `json:"coalesced"`
	Requeued   uint64 `json:"requeued"`
	Rejected   uint64 `json:"rejected"`
	Stopped    uint64 `json:"stopped"`
	Started    uint64 `json:"started"`
	Completed  uint64 `json:"completed"`
	ScanErrors uint64 `json:"scan_errors"`
	Malicious  uint64 `json:"malicious"`
	Abandoned  uint64 `json:"abandoned"`
	Active     int    `json:"active"`
	Queued     int    `json:"queued"`
	QueueLimit int    `json:"queue_limit"`
	Capacity   int    `json:"capacity"`
}

type pendingEvent struct {
	event   FileEvent
	running bool
	dirty   bool
}

type Orchestrator struct {
	sc         FileScanner
	ex         *Excluder
	workers    int
	queue      chan *pendingEvent
	queueLimit int
	wg         sync.WaitGroup

	// onMalicious, if set, is invoked with the path+signature of a file the
	// worker just found malicious. The service wires it to terminate any running
	// process launched from that image, so a malicious executable that is dropped
	// AND immediately run is both quarantined (file) and killed (process) — the
	// file-watcher scan otherwise only removes the on-disk image while the process
	// keeps running (and can outrace the process guard, which loses its scan
	// target once the image is quarantined).
	onMalicious func(path, signature string)

	// Queued duplicates share one future snapshot. A notification received after
	// scanning starts marks the path dirty: its later bytes need a follow-up.
	mu       sync.Mutex
	inflight map[string]*pendingEvent
	ctx      context.Context
	stats    Stats
}

func New(sc FileScanner, ex *Excluder, workers, queueSize int) *Orchestrator {
	if workers < 1 {
		workers = 1
	}
	if queueSize < 1 {
		queueSize = 1
	}
	return &Orchestrator{
		sc:      sc,
		ex:      ex,
		workers: workers,
		// Reserve one return slot per worker so dirty work can join the tail
		// even when arrivals filled the admission queue. Workers never block
		// each other returning work, and one hot path cannot monopolize them.
		queue:      make(chan *pendingEvent, queueSize+workers),
		queueLimit: queueSize,
		inflight:   make(map[string]*pendingEvent),
	}
}

// SetOnMalicious registers a callback invoked with the path+signature of a file
// found malicious by a worker (used to kill processes running that image).
func (o *Orchestrator) SetOnMalicious(fn func(path, signature string)) { o.onMalicious = fn }

func (o *Orchestrator) Start(ctx context.Context) {
	o.mu.Lock()
	o.ctx = ctx
	o.mu.Unlock()
	for i := 0; i < o.workers; i++ {
		o.wg.Add(1)
		go o.worker(ctx)
	}
}

func (o *Orchestrator) worker(ctx context.Context) {
	defer o.wg.Done()
	for {
		if ctx.Err() != nil {
			return
		}
		select {
		case <-ctx.Done():
			return
		case pending := <-o.queue:
			if ctx.Err() != nil {
				return
			}
			o.mu.Lock()
			pending.running = true
			ev := pending.event
			o.stats.Active++
			o.stats.Started++
			o.mu.Unlock()
			var scanErr error
			var malicious bool
			if o.ex == nil || !o.ex.Excluded(ev.Path) {
				// Errors (file vanished, engine down) are non-fatal; the file stays
				// unknown and will be revisited on the next event/sweep.
				source := ev.Source
				if source == "" {
					source = event.SourceFileWatcher
				}
				verdict, sig, err := o.sc.ScanFile(ev.Path, source)
				scanErr = err
				malicious = verdict == cache.VerdictMalicious
				if verdict == cache.VerdictMalicious && o.onMalicious != nil {
					o.onMalicious(ev.Path, sig)
				}
			}
			o.mu.Lock()
			o.stats.Active--
			o.stats.Completed++
			if scanErr != nil {
				o.stats.ScanErrors++
			}
			if malicious {
				o.stats.Malicious++
			}
			pending.running = false
			if pending.dirty && ctx.Err() == nil {
				pending.dirty = false
				o.stats.Requeued++
				// At most queueLimit+workers paths are admitted, and this
				// worker's path is outside the queue: its reserved slot is free.
				o.queue <- pending
			} else {
				if pending.dirty {
					o.stats.Abandoned++
				}
				delete(o.inflight, ev.Path)
			}
			o.mu.Unlock()
		}
	}
}

// Enqueue returns false for a full admission queue or canceled worker context.
// Duplicates waiting to scan coalesce; duplicates during analysis retain one
// follow-up at the queue tail. Further changes during that scan retain another.
func (o *Orchestrator) Enqueue(ev FileEvent) bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.ctx != nil && o.ctx.Err() != nil {
		o.stats.Stopped++
		return false
	}
	if pending := o.inflight[ev.Path]; pending != nil {
		pending.event = ev
		pending.dirty = pending.dirty || pending.running
		o.stats.Coalesced++
		return true
	}
	if len(o.queue) >= o.queueLimit {
		o.stats.Rejected++
		return false
	}
	pending := &pendingEvent{event: ev}
	o.inflight[ev.Path] = pending
	o.queue <- pending
	o.stats.Enqueued++
	return true
}

func (o *Orchestrator) Snapshot() Stats {
	o.mu.Lock()
	defer o.mu.Unlock()
	s := o.stats
	s.Queued = len(o.inflight) - s.Active
	s.QueueLimit, s.Capacity = o.queueLimit, cap(o.queue)
	return s
}

// Stop waits for active scans after cancellation, then releases queued paths.
func (o *Orchestrator) Stop() {
	o.wg.Wait()
	o.mu.Lock()
	defer o.mu.Unlock()
	o.stats.Abandoned += uint64(len(o.inflight))
	clear(o.inflight)
	for len(o.queue) > 0 {
		<-o.queue
	}
}
