package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/go-sdk/v2/plugins"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	done := startQueue(ctx)
	defer func() {
		cancel()
		<-done
	}()

	err := plugins.InitAnalysisPlugin("com.utmstack.events", analyze)
	if err != nil {
		_ = catcher.Error("failed to start analysis plugin", err, map[string]any{
			"process": "plugin_com.utmstack.events",
		})
		time.Sleep(5 * time.Second)
		cancel()
		<-done
		os.Exit(1)
	}
}

func analyze(event *plugins.Event, _ plugins.Analysis_AnalyzeServer) error {
	if event == nil {
		return catcher.Error("cannot convert event to json", nil, map[string]any{"process": "plugin_com.utmstack.events"})
	}
	jLog, err := plugins.ProtoMessageToString(event)
	if err != nil {
		return catcher.Error("cannot convert event to json", err, map[string]any{"process": "plugin_com.utmstack.events"})
	}

	addToQueue(*jLog)

	return io.EOF
}
