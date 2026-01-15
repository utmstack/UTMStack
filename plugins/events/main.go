package main

import (
	"io"
	"os"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/threatwinds/go-sdk/utils"
)

func main() {
	startQueue()

	err := plugins.InitAnalysisPlugin("com.utmstack.events", analyze)
	if err != nil {
		_ = catcher.Error("failed to start analysis plugin", err, map[string]any{
			"process": "plugin_com.utmstack.events",
		})
		os.Exit(1)
	}
}

func analyze(event *plugins.Event, _ plugins.Analysis_AnalyzeServer) error {
	jLog, err := utils.ProtoMessageToString(event)
	if err != nil {
		return catcher.Error("cannot convert event to json", err, map[string]any{"process": "plugin_com.utmstack.events"})
	}

	addToQueue(*jLog)

	return io.EOF
}
