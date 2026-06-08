package main

import (
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/plugins/soar/internal/cache"
	"github.com/utmstack/UTMStack/plugins/soar/internal/client"
	"github.com/utmstack/UTMStack/plugins/soar/internal/dispatcher"
	"github.com/utmstack/UTMStack/plugins/soar/internal/handler"
)

const pluginID = "com.utmstack.soar"

func main() {
	if plugins.GetCfg("plugin_" + pluginID).GetEnv().Mode == "playground" {
		return
	}

	backend := client.New()
	cache.Init(backend)

	go cache.RefreshLoop()
	go dispatcher.Start(backend)

	if err := plugins.InitCorrelationPlugin(pluginID, handler.New(backend)); err != nil {
		_ = catcher.Error(pluginID, err, map[string]any{"process": "plugin_" + pluginID})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}
