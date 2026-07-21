package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/plugins/enrichment/config"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/parselog"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/storage"
	"github.com/utmstack/UTMStack/plugins/enrichment/internal/watcher"
)

func main() {
	if err := storage.SyncFromDisk(); err != nil {
		_ = catcher.Error("initial sync from disk failed", err, map[string]any{
			"process": config.ProcessName,
		})
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go watcher.Start(ctx)

	err := plugins.InitParsingPlugin(config.PluginName, parselog.ParseLog)
	if err != nil {
		_ = catcher.Error("failed to init parsing plugin", err, map[string]any{
			"process": config.ProcessName,
		})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}
