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

	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/upstream"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
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
	utils.Logger.Info("Stopping UTMStack Forwarder...")

	p.mu.Lock()
	cancel := p.cancel
	p.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.Logger.Info("All goroutines stopped gracefully")
	case <-time.After(shutdownTimeout):
		utils.Logger.ErrorF("Shutdown timeout after %v", shutdownTimeout)
	}

	collectorpkg.StopAll()
	utils.Logger.Info("UTMStack Forwarder stopped")
	return nil
}

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
	utils.InitLogger(config.CollectorLogFile)

	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	p.mu.Lock()
	p.cancel = cancel
	p.mu.Unlock()

	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.CollectorKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.CollectorID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "collector")

	p.goSafe("StartPing", func() {
		upstream.StartPing(cnf, ctx)
	})
	p.goSafe("StartCollectorConfigStream", func() {
		upstream.StartCollectorConfigStream(cnf, ctx)
	})
	p.goSafe("ProcessLogs", func() {
		upstream.ProcessLogs(cnf, ctx, collectorpkg.LogQueue)
	})

	if err := collectorpkg.SyncCollectorConfig(); err != nil {
		utils.Logger.ErrorF("error syncing collector config: %v", err)
	}

	collectorpkg.StartAll(ctx)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-signals:
		utils.Logger.Info("Received signal: %v", sig)
	case <-ctx.Done():
		utils.Logger.Info("Context cancelled")
	}
}
