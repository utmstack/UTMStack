package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const pluginName = "com.utmstack.ad-audit"

var cacheMu = sync.Mutex{}

func main() {
	loadBackendConfig()
	seedCacheFromBackend()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go flushLoop(ctx, 60*time.Second)

	if err := plugins.InitAnalysisPlugin(pluginName, analyze); err != nil {
		_ = catcher.Error("failed to start analysis plugin", err, map[string]any{"process": "plugin_" + pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

const (
	wineventlogDataType = "wineventlog"
	linuxDataType       = "linux"
)

func analyze(event *plugins.Event, _ plugins.Analysis_AnalyzeServer) error {
	switch event.GetDataType() {
	case wineventlogDataType:
		return handleWindows(event)
	case linuxDataType:
		return handleLinux(event)
	default:
		return io.EOF
	}
}
