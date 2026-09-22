package main

import (
	"github.com/utmstack/UTMStack/collectors/forwarder/cmd"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

func main() {
	utils.InitLogger(config.CollectorLogFile)
	cmd.Execute()
}
