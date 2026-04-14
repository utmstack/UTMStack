package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/scheduler"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/workers"
)

func main() {
	catcher.Info("Starting Compliance Orchestrator", map[string]any{
		"process": "compliance-orchestrator",
	})

	backend, err := bootstrap()
	if err != nil {
		catcher.Info("Compliance Orchestrator bootstrap failed", map[string]any{})
		return
	}

	catcher.Info("Compliance Orchestrator bootstrapped successfully", nil)

	ctx := context.Background()

	go workers.StartWorkers(ctx, backend)

	go scheduler.StartScheduler(ctx, backend)

	for {
		time.Sleep(1 * time.Hour)
	}
}
