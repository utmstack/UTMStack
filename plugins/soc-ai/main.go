package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/api"
	"github.com/utmstack/UTMStack/plugins/soc-ai/internal/queue"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	if plugins.GetCfg("plugin_com.utmstack.soc-ai").GetEnv().Mode == "playground" {
		return
	}

	go config.StartConfigurationSystem()

	time.Sleep(config.CONFIG_STARTUP_DELAY * time.Second)
	queue.Initialize()

	// Start HTTP API server for manual alert submission
	go api.StartHTTPServer()

	err := plugins.InitCorrelationPlugin("com.utmstack.soc-ai", correlate)
	if err != nil {
		_ = catcher.Error("failed to start correlation plugin", err, map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
		})
		time.Sleep(config.ERROR_EXIT_DELAY * time.Second)
		os.Exit(1)
	}
}

func correlate(_ context.Context, alert *plugins.Alert) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate method", nil, map[string]any{
				"process": "plugin_com.utmstack.soc-ai",
				"panic":   r,
				"alert":   alert.Name,
			})
		}
	}()

	cfg := config.GetConfig()
	if cfg == nil || !cfg.ModuleActive {
		return &emptypb.Empty{}, nil
	}

	// Skip automatic analysis if AutoAnalyze is disabled
	// Manual submissions via HTTP API are not affected by this setting
	if !cfg.AutoAnalyze {
		return &emptypb.Empty{}, nil
	}

	if !queue.Enqueue(alert) {
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}
