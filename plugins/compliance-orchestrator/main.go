package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/scheduler"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/workers"
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.compliance-orchestrator").Env.Mode
	if mode != "manager" {
		return
	}

	catcher.Info("Starting Compliance Orchestrator", map[string]any{
		"process": "plugin_com.utmstack.compliance-orchestrator",
	})

	backend, err := bootstrap()
	if err != nil {
		_ = catcher.Error("Compliance Orchestrator bootstrap failed", err, map[string]any{
			"process": "plugin_com.utmstack.compliance-orchestrator",
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	catcher.Info("Compliance Orchestrator bootstrapped successfully", map[string]any{
		"process": "plugin_com.utmstack.compliance-orchestrator",
	})

	ctx := context.Background()

	go workers.StartWorkers(ctx, backend)

	go scheduler.StartScheduler(ctx, backend)

	for {
		time.Sleep(1 * time.Hour)
	}
}
