package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/utmstack/UTMStack/edr/internal/daemon"
)

func main() {
	cfg, err := daemon.FromEnv(os.Getenv)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Printf("EDR mirror daemon starting (mirror=%s)", cfg.MirrorDir)
	if err := daemon.Run(ctx, cfg); err != nil {
		log.Fatalf("run: %v", err)
	}
}
