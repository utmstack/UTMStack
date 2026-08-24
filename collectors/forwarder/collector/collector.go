package collector

import (
	"context"
	"sync"

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

var (
	activeCollectors []Collector
	collectorsMu     sync.Mutex

	// LogQueue is the shared channel that all collectors push logs to.
	LogQueue chan *plugins.Log
)

func init() {
	LogQueue = make(chan *plugins.Log, 1000)
}

// StartAll starts all collectors: syslog, netflow, file, and http.
// The context is used for graceful shutdown signaling.
func StartAll(ctx context.Context) {
	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	// Clear previous collectors
	activeCollectors = nil

	// Create syslog collector
	syslogCollector := syslog.New()
	activeCollectors = append(activeCollectors, syslogCollector)
	go runCollector(ctx, syslogCollector, LogQueue)

	// Create netflow collector
	netflowCollector := netflow.New()
	activeCollectors = append(activeCollectors, netflowCollector)
	go runCollector(ctx, netflowCollector, LogQueue)

	// Create file collector (nginx, postgresql, etc.)
	fileCollector := file.New()
	activeCollectors = append(activeCollectors, fileCollector)
	go runCollector(ctx, fileCollector, LogQueue)

	// Create HTTP/HTTPS collector
	httpColl := http.New()
	activeCollectors = append(activeCollectors, httpColl)
	go runCollector(ctx, httpColl, LogQueue)

	utils.Logger.Info("All collectors started")
}

// runCollector runs a collector with panic recovery.
func runCollector(ctx context.Context, c Collector, queue chan *plugins.Log) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in collector %s: %v", c.Name(), r)
		}
	}()
	c.Start(ctx, queue)
}

// StopAll stops all active collectors.
func StopAll() {
	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	for _, c := range activeCollectors {
		utils.Logger.Info("Stopping collector: %s", c.Name())
		c.Stop()
	}
	activeCollectors = nil
	utils.Logger.Info("All collectors stopped")
}
