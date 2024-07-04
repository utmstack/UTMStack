package main

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/self/config"
	"github.com/utmstack/UTMStack/agent/self/update"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitLogger(filepath.Join(path, "logs", config.SERV_LOG))

	utils.SelfLogger.Info("Updating %s...", config.SERV_NAME)

	if isRunning, err := utils.CheckIfServiceIsActive(config.SERV_NAME); err != nil {
		utils.SelfLogger.Fatal("error checking %s service: %v", config.SERV_NAME, err)
	} else if isRunning {
		err = utils.StopService(config.SERV_NAME)
		if err != nil {
			utils.SelfLogger.Fatal("error stopping %s service: %v", config.SERV_NAME, err)
		}
		utils.SelfLogger.Info("%s stopped correctly", config.SERV_NAME)
	}

	err := update.UpdateService()
	if err != nil {
		utils.SelfLogger.Fatal("error updating new %s service: %v", config.SERV_NAME, err)
	}
	utils.SelfLogger.Info("New %s downloaded correctly", config.SERV_NAME)

	err = utils.RestartService(config.SERV_NAME)
	if err != nil {
		utils.SelfLogger.Fatal("error restarting %s service: %v", config.SERV_NAME, err)
	}
	utils.SelfLogger.Info("%s updated correctly", config.SERV_NAME)
}
