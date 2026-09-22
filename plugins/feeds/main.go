package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/client"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/initializer"
)

const (
	readinessInterval = 5 * time.Second
	readinessLogEvery = 60 * time.Second
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.feeds").Env.Mode
	if mode != "manager" {
		return
	}

	catcher.Info("Starting ThreadWinds Ingestion Service", nil)

	// What the backend decided for this plugin — read before anything asks for
	// it, and kept current from there on.
	config.StartConfigurationSystem()

	ic := &client.CustomersManagerClient{}
	if err := ic.LoadInstanceConfig(); err != nil {
		_ = catcher.Error("instance configuration not loaded within deadline", err, nil)
		os.Exit(1)
	}

	waitForCMReachable(ic.Server)

	ctx := context.Background()
	app, err := initializer.NewApp(ctx, ic)
	if err != nil {
		_ = catcher.Error("failed to initialize application", err, nil)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go app.Run(ctx)

	sig := <-sigChan
	catcher.Info("received shutdown signal, initiating graceful shutdown", map[string]any{
		"signal": sig.String(),
	})

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		_ = catcher.Error("error during shutdown", err, nil)
		time.Sleep(5 * time.Second)
	}

	catcher.Info("ThreadWinds Ingestion Service stopped", nil)
}

func waitForCMReachable(server string) {
	url := fmt.Sprintf("%s/proxy/usage", server)
	httpClient := &http.Client{Timeout: 5 * time.Second}

	var lastLog time.Time
	for {
		resp, err := httpClient.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			catcher.Info("CM server is reachable, starting", map[string]any{
				"server": server,
				"status": resp.StatusCode,
			})
			return
		}

		if time.Since(lastLog) >= readinessLogEvery || lastLog.IsZero() {
			_ = catcher.Error("CM server not reachable yet", err, map[string]any{
				"server":   server,
				"retry_in": readinessInterval.String(),
			})
			lastLog = time.Now()
		}
		time.Sleep(readinessInterval)
	}
}
