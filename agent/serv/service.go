package serv

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/threatwinds/go-sdk/plugins"

	pb "github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/collector/auditd"
	"github.com/utmstack/UTMStack/agent/collector/platform"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/database"
	"github.com/utmstack/UTMStack/agent/dependency"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
	"google.golang.org/grpc/metadata"
)

const shutdownTimeout = 30 * time.Second

type program struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex
}

func (p *program) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(_ service.Service) error {
	utils.Logger.Info("Stopping UTMStack Agent...")

	p.mu.Lock()
	cancel := p.cancel
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	// Stop all OS collectors
	stopOSCollectors()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.Logger.Info("All goroutines stopped gracefully")
	case <-time.After(shutdownTimeout):
		utils.Logger.ErrorF("Shutdown timeout after %v, some goroutines may not have stopped", shutdownTimeout)
	}

	// Close database
	if db, err := database.GetDB(); err == nil && db != nil {
		if err := db.Close(); err != nil {
			utils.Logger.ErrorF("error closing database: %v", err)
		}
	}

	utils.Logger.Info("UTMStack Agent stopped")
	return nil
}

// goSafe launches a goroutine with panic recovery and WaitGroup tracking.
func (p *program) goSafe(name string, fn func()) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() {
			if r := recover(); r != nil {
				utils.Logger.ErrorF("panic in %s: %v", name, r)
			}
		}()
		fn()
	}()
}

func (p *program) run() {
	utils.InitLogger(config.ServiceLogFile)
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	db, err := database.GetDB()
	if err != nil {
		utils.Logger.ErrorF("error initializing database: %v", err)
	} else if err = db.Migrate(models.Log{}); err != nil {
		utils.Logger.ErrorF("error migrating logs table: %v", err)
	}

	// Reconcile dependencies (updater, beats, etc.) before starting collectors
	if err := dependency.Reconcile(cnf.Server, cnf.SkipCertValidation); err != nil {
		utils.Logger.ErrorF("error reconciling dependencies: %v", err)
		// Continue anyway - agent should try to run with what it has
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.mu.Lock()
	p.cancel = cancel
	p.mu.Unlock()

	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")

	// Start all goroutines with panic recovery
	p.goSafe("IncidentResponseStream", func() {
		pb.IncidentResponseStream(cnf, ctx)
	})

	p.goSafe("StartPing", func() {
		pb.StartPing(cnf, ctx)
	})

	p.goSafe("ProcessLogs", func() {
		logProcessor, err := pb.GetLogProcessor()
		if err != nil {
			utils.Logger.ErrorF("error initializing log processor: %v", err)
			return
		}
		logProcessor.ProcessLogs(cnf, ctx)
	})

	p.goSafe("UpdateAgent", func() {
		pb.UpdateAgent(cnf, ctx)
	})

	// Start OS-level collectors (platform + auditd only).
	startOSCollectors(ctx, pb.LogQueue)

	// Wait for shutdown signal
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signals:
		utils.Logger.Info("Received signal: %v", sig)
	case <-ctx.Done():
		utils.Logger.Info("Context cancelled")
	}
}

var activeCollectors []platform.Collector

func startOSCollectors(ctx context.Context, queue chan *plugins.Log) {
	activeCollectors = nil
	for _, c := range platform.GetCollectors() {
		activeCollectors = append(activeCollectors, c)
		go runCollector(ctx, c, queue)
	}
	a := auditd.New()
	activeCollectors = append(activeCollectors, a)
	go runCollector(ctx, a, queue)
}

func stopOSCollectors() {
	for _, c := range activeCollectors {
		c.Stop()
	}
	activeCollectors = nil
}

func runCollector(ctx context.Context, c platform.Collector, queue chan *plugins.Log) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in collector %s: %v", c.Name(), r)
		}
	}()
	c.Start(ctx, queue)
}

