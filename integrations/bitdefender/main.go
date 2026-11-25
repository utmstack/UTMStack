package main

import (
	"runtime"

	"github.com/utmstack/go-sdk/v2/plugins"

	"github.com/utmstack/UTMStack/integrations/bitdefender/config"
	"github.com/utmstack/UTMStack/integrations/bitdefender/server"
)

func main() {
	go config.StartConfigurationSystem()

	for t := 0; t < 2*runtime.NumCPU(); t++ {
		go func() {
			plugins.SendLogsFromChannel("com.utmstack.bitdefender")
		}()
	}

	server.StartServer()
}
