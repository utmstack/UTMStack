package watcher

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/UTMStack/plugins/enrichment/config"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/storage"
)

func Start(ctx context.Context) {
	ticker := time.NewTicker(config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			catcher.Info("watcher stopped", map[string]any{
				"process": config.ProcessName,
			})
			return
		case <-ticker.C:
			if err := storage.SyncFromDisk(); err != nil {
				_ = catcher.Error("sync from disk failed", err, map[string]any{
					"process": config.ProcessName,
				})
			}
		}
	}
}
