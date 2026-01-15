package main

import (
	"runtime"

	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/plugins/bitdefender/config"
	"github.com/utmstack/UTMStack/plugins/bitdefender/server"
)

func main() {
	mode := plugins.GetCfg("plugin_com.utmstack.bitdefender").Env.Mode
	if mode != "manager" {
		return
	}

	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.bitdefender")
		}()
	}

	server.StartServer()
}
