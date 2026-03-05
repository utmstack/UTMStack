package main

import (
	"github.com/utmstack/UTMStack/agent/cmd"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)
	cmd.Execute()
}
