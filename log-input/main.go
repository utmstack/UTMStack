package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/UTMStack/log-input/auth"
	"github.com/utmstack/UTMStack/log-input/config"
	"github.com/utmstack/UTMStack/log-input/ingest"
	"github.com/utmstack/UTMStack/log-input/publish"
)

const processName = "log-input"

func main() {
	cfg, err := config.Load()
	if err != nil {
		_ = catcher.Error("invalid configuration", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	authService := auth.New(cfg)
	if err := authService.Ping(ctx); err != nil {
		_ = catcher.Error("cannot reach the auth cache", err, map[string]any{
			"process": processName,
			"redis":   cfg.RedisAddr,
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer func() { _ = authService.Close() }()

	publisher, err := publish.New(ctx, cfg.NATSURL, cfg.Shards)
	if err != nil {
		_ = catcher.Error("cannot reach the log queue", err, map[string]any{
			"process": processName,
			"nats":    cfg.NATSURL,
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	defer publisher.Close()

	server, err := ingest.NewServer(cfg, authService, publisher)
	if err != nil {
		_ = catcher.Error("cannot build the server", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	httpServer, err := ingest.NewHTTPServer(cfg, authService, publisher)
	if err != nil {
		_ = catcher.Error("cannot build the http ingest server", err, map[string]any{"process": processName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	healthEndpoint := ingest.NewHealth(cfg.HealthAddr)
	go healthEndpoint.Serve()

	errs := make(chan error, 1)
	go func() { errs <- server.Serve() }()
	go func() {
		if err := httpServer.Serve(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = catcher.Error("the http ingest server stopped", err, map[string]any{"process": processName})
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errs:
		if err != nil {
			_ = catcher.Error("the server stopped", err, map[string]any{"process": processName})
			time.Sleep(5 * time.Second)
			os.Exit(1)
		}
	case <-signals:
		catcher.Info("shutting down", map[string]any{"process": processName})
		healthEndpoint.Draining()
		server.Stop()

		sCtx, sCancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		_ = httpServer.Stop(sCtx)
		healthEndpoint.Close(sCtx)
		sCancel()
	}
}
