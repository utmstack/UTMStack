package main

import (
	"fmt"
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

var (
	InternalKey    string
	BackendService string
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.modules-config").Env.Mode
	if mode != "manager" {
		return
	}

	utmConfig := plugins.PluginCfg("com.utmstack")
	InternalKey = utmConfig.Get("internalKey").String()
	BackendService = utmConfig.Get("backend").String()

	if InternalKey == "" || BackendService == "" {
		_ = catcher.Error("error getting configuration", fmt.Errorf("internal key or backend service is empty"), map[string]any{"process": "plugin_com.utmstack.modules-config"})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	go startGRPCServer()
	startHTTPServer()
}
