package main

import (
	"fmt"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

var (
	InternalKey    string
	BackendService string
)

func main() {
	mode := plugins.GetCfg().Env.Mode
	if mode != "manager" {
		return
	}

	utmConfig := plugins.PluginCfg("com.utmstack", false)
	InternalKey = utmConfig.Get("internalKey").String()
	BackendService = utmConfig.Get("backend").String()

	if InternalKey == "" || BackendService == "" {
		_ = catcher.Error("error getting configuration", fmt.Errorf("internal key or backend service is empty"), nil)
		return
	}

	go startGRPCServer()
	startHTTPServer()
}
