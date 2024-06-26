package main

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/self/configuration"
	"github.com/utmstack/UTMStack/agent/self/update"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func main() {
	path := utils.GetMyPath()
	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))

	h.Info("Updating UTMStackAgent...")

	if isRunning, err := utils.CheckIfServiceIsActive(configuration.SERV_NAME); err != nil {
		h.Fatal("error checking %s service: %v", configuration.SERV_NAME, err)
	} else if isRunning {
		err = utils.StopService(configuration.SERV_NAME)
		if err != nil {
			h.Fatal("error stopping %s service: %v", configuration.SERV_NAME, err)
		}
		h.Info("%s stopped correctly", configuration.SERV_NAME)
	}

	err := update.UpdateService()
	if err != nil {
		h.Fatal("error updating new %s service: %v", configuration.SERV_NAME, err)
	}
	h.Info("New %s downloaded correctly", configuration.SERV_NAME)

	err = utils.RestartService(configuration.SERV_NAME)
	if err != nil {
		h.Fatal("error restarting %s service: %v", configuration.SERV_NAME, err)
	}
	h.Info("%s updated correctly", configuration.SERV_NAME)
}
