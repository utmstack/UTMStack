package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
	"google.golang.org/protobuf/types/known/emptypb"
)

type socAiServer struct {
	plugins.UnimplementedCorrelationServer
}

func main() {
	utils.Logger.Info("Starting soc-ai plugin...")

	go config.StartConfigurationSystem()

	time.Sleep(2 * time.Second)
	InitializeQueue()

	_ = plugins.InitCorrelationPlugin("com.utmstack.soc-ai", correlate)
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
		utils.Logger.LogF(100, "SOC-AI module is disabled, skipping alert: %s", alert.Id)
		return &emptypb.Empty{}, nil
	}

	if !EnqueueAlert(alert) {
		utils.Logger.LogF(300, "Alert %s was dropped due to full queue", alert.Id)
		return &emptypb.Empty{}, nil
	}

	return &emptypb.Empty{}, nil
}
