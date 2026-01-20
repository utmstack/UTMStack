package main

import (
	"context"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	if plugins.GetCfg("plugin_com.utmstack.soc-ai").GetEnv().Mode == "playground" {
		return
	}

	go config.StartConfigurationSystem()

	time.Sleep(2 * time.Second)
	initializeQueue()

	err := plugins.InitCorrelationPlugin("com.utmstack.soc-ai", correlate)
	if err != nil {
		_ = catcher.Error("failed to start correlation plugin", err, map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
		})
		os.Exit(1)
	}
}

func correlate(_ context.Context,
	alert *plugins.Alert) (*emptypb.Empty, error) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in Correlate method", nil, map[string]any{
				"process": "plugin_com.utmstack.soc-ai",
				"panic":   r,
				"alert":   alert.Name,
			})
		}
	}()

	// Check if the module is active before processing the alert
	if config.GetConfig() == nil || !config.GetConfig().ModuleActive {
		return &emptypb.Empty{}, nil
	}

	if !enqueueAlert(alert) {
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}
