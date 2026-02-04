package scheduler

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/client"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/models"
)

var Jobs chan models.ControlConfig

func StartScheduler(ctx context.Context, backend *client.BackendClient) {
	Jobs = make(chan models.ControlConfig, 1000)

	ticker := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-ctx.Done():
			// Shutdown limpio
			return

		case <-ticker.C:
			configs, err := backend.GetControlConfigs(ctx)
			if err != nil {
				continue
			}

			catcher.Info("Scheduler: sending configs", map[string]any{
				"cantidad":  len(configs),
				"timestamp": time.Now().String(),
			})

			for _, cfg := range configs {
				Jobs <- cfg
			}
		}
	}
}
