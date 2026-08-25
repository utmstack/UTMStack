package collector

import (
	"context"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/file"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/http"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/netflow"
	"github.com/utmstack/UTMStack/collectors/forwarder/collector/syslog"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

// Collector is the interface that every collector must implement.
type Collector interface {
	Name() string
	Start(ctx context.Context, queue chan *plugins.Log)
	Stop()
}

const watchQuiesceTimeout = 10 * time.Second

type collectorHandle struct {
	collector Collector
	done      chan struct{}
}

var (
	activeCollectors []collectorHandle
	collectorsMu     sync.Mutex
	cancelWatchers   context.CancelFunc
	LogQueue         chan *plugins.Log
)

func init() {
	LogQueue = make(chan *plugins.Log, 1000)
}

func StartAll(ctx context.Context) {
	startAll(ctx, LogQueue,
		syslog.New(),
		netflow.New(),
		file.New(),
		http.New(),
	)
}

func startAll(ctx context.Context, queue chan *plugins.Log, collectors ...Collector) {
	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	// Clear previous collectors
	activeCollectors = nil

	watchCtx, cancel := context.WithCancel(ctx)
	cancelWatchers = cancel

	for _, c := range collectors {
		handle := collectorHandle{collector: c, done: make(chan struct{})}
		activeCollectors = append(activeCollectors, handle)
		go runCollector(watchCtx, c, queue, handle.done)
	}

	utils.Logger.Info("All collectors started")
}

func runCollector(ctx context.Context, c Collector, queue chan *plugins.Log, done chan struct{}) {
	defer close(done)
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in collector %s: %v", c.Name(), r)
		}
	}()
	c.Start(ctx, queue)
}

// StopAll stops all active collectors.
func StopAll() {
	stopAll(watchQuiesceTimeout)
}

func stopAll(timeout time.Duration) {
	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	if cancelWatchers != nil {
		cancelWatchers()
		cancelWatchers = nil
	}

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	timedOut := false

	// Wait for every watch loop to quiesce before stopping anything, so no
	// reconcile can reopen a socket behind Stop. The timer is shared, so the
	// bound covers this whole phase instead of each collector separately.
	//
	// On expiry the ordering guarantee no longer holds: Stop then runs on a
	// collector whose reconcile may still be in flight, which is exactly the
	// race this phase exists to prevent. Doing it anyway is deliberate, because
	// leaving a collector alive is worse, but it is a degraded path.
	for _, handle := range activeCollectors {
		// Check without blocking first. A collector that already returned must
		// never be reported as late, and it could be if the timer is also ready,
		// because select picks at random when several cases are ready.
		select {
		case <-handle.done:
			continue
		default:
		}

		if timedOut {
			utils.Logger.ErrorF("collector %s did not stop watching within %v, stopping it anyway", handle.collector.Name(), timeout)
			continue
		}

		select {
		case <-handle.done:
		case <-timer.C:
			timedOut = true
			// done may have closed while we were blocked here; confirm before
			// reporting this collector as late.
			select {
			case <-handle.done:
			default:
				utils.Logger.ErrorF("collector %s did not stop watching within %v, stopping it anyway", handle.collector.Name(), timeout)
			}
		}
	}

	for _, handle := range activeCollectors {
		utils.Logger.Info("Stopping collector: %s", handle.collector.Name())
		handle.collector.Stop()
	}
	activeCollectors = nil

	if timedOut {
		utils.Logger.ErrorF("all collectors stopped, but at least one did not quiesce within %v", timeout)
		return
	}
	utils.Logger.Info("All collectors stopped")
}
