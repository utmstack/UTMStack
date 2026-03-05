package collector

import (
	"context"
	"fmt"
	"sync"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/collector/file"
	"github.com/utmstack/UTMStack/agent/collector/netflow"
	"github.com/utmstack/UTMStack/agent/collector/platform"
	"github.com/utmstack/UTMStack/agent/collector/syslog"
	"github.com/utmstack/UTMStack/agent/utils"
)

// Collector is the interface that every collector must implement.
type Collector interface {
	Name() string
	Start(ctx context.Context, queue chan *plugins.Log)
	Stop()
}

// Installable is an optional interface for collectors that install system services.
type Installable interface {
	Install() error
	Uninstall() error
}

var (
	activeCollectors []Collector
	collectorsMu     sync.Mutex
)

// StartAll starts all collectors: platform-specific, syslog and netflow.
// The context is used for graceful shutdown signaling.
func StartAll(ctx context.Context) {
	collectorsMu.Lock()
	defer collectorsMu.Unlock()

	// Clear previous collectors
	activeCollectors = nil

	// Create syslog collector
	syslogCollector := syslog.New()
	activeCollectors = append(activeCollectors, syslogCollector)
	go runCollector(ctx, syslogCollector, agent.LogQueue)

	// Create netflow collector
	netflowCollector := netflow.New()
	activeCollectors = append(activeCollectors, netflowCollector)
	go runCollector(ctx, netflowCollector, agent.LogQueue)

	// Create file collector (nginx, postgresql, etc.)
	fileCollector := file.New()
	activeCollectors = append(activeCollectors, fileCollector)
	go runCollector(ctx, fileCollector, agent.LogQueue)

	// Create platform collectors (filebeat, winlogbeat, macos, etc.)
	platformCollectors := platform.GetCollectors()
	for _, c := range platformCollectors {
		activeCollectors = append(activeCollectors, c)
		go runCollector(ctx, c, agent.LogQueue)
	}

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

// InstallAll installs all platform collectors that implement Installable.
func InstallAll() error {
	platformCollectors := platform.GetCollectors()
	for _, c := range platformCollectors {
		if inst, ok := c.(Installable); ok {
			if err := inst.Install(); err != nil {
				return fmt.Errorf("%v", err)
			}
		}
	}
	utils.Logger.LogF(100, "collectors installed correctly")
	return nil
}

// UninstallAll uninstalls all platform collectors that implement Installable.
func UninstallAll() error {
	platformCollectors := platform.GetCollectors()
	for _, c := range platformCollectors {
		if inst, ok := c.(Installable); ok {
			if err := inst.Uninstall(); err != nil {
				return fmt.Errorf("%v", err)
			}
		}
	}
	utils.Logger.LogF(100, "collectors uninstalled correctly")
	return nil
}
