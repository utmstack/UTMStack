package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/initializer"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/utils"
)

const (
	urlCheckConnection = "https://apis.threatwinds.com"
)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		return
	}

	catcher.Info("Starting ThreadWinds Ingestion Service", nil)

	for {
		if err := utils.ConnectionChecker(urlCheckConnection); err != nil {
			_ = catcher.Error("External connection failure detected", err, nil)
			continue
		}
		break
	}

	ctx := context.Background()
	app, err := initializer.NewApp(ctx)
	if err != nil {
		catcher.Error("failed to initialize application", err, nil)
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
		catcher.Error("error during shutdown", err, nil)
	}

	catcher.Info("ThreadWinds Ingestion Service stopped", nil)
}
