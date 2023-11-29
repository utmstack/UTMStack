package main

import (
	"log"
	"path/filepath"

	"github.com/KbaYero/UTMStack/agent/self/configuration"
	"github.com/KbaYero/UTMStack/agent/self/constants"
	"github.com/KbaYero/UTMStack/agent/self/update"
	"github.com/KbaYero/UTMStack/agent/self/utils"
	"github.com/quantfall/holmes"
)

var h = holmes.New("debug", constants.SERV_NAME)

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		h.FatalError("Failed to get current path: %v", err)
	}

	// Configuring log saving
	var logger = utils.CreateLogger(filepath.Join(path, "logs", constants.SERV_LOG))
	defer logger.Close()
	log.SetOutput(logger)

	// Save data from versions.json
	allVersions := update.Version{}
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &allVersions)
	if err != nil {
		h.FatalError("error reading current versions.json: %v", err)
	}

	// Select environment
	env, err := configuration.ReadEnv()
	if err != nil {
		h.FatalError("Failed to get current path: %v", err)
	}

	h.Info("Updating UTMStackUpdater...")

	if isRunning, err := utils.CheckIfServiceIsActive(constants.UPDATER_SERV_NAME); err != nil {
		h.FatalError("error checking %s service: %v", constants.UPDATER_SERV_NAME, err)
	} else if isRunning {
		err = utils.StopService(constants.UPDATER_SERV_NAME)
		if err != nil {
			h.FatalError("error stopping %s service: %v", constants.UPDATER_SERV_NAME, err)
		}
		h.Info("UTMStackUpdater stopped correctly")
	}

	err = update.UpdateUpdaterService(allVersions.UpdaterVersion, env.Branch)
	if err != nil {
		h.FatalError("error downloading new %s service: %v", constants.UPDATER_SERV_NAME, err)
	}
	h.Info("New UTMStackUpdater downloaded correctly")

	err = utils.RestartService(constants.UPDATER_SERV_NAME)
	if err != nil {
		h.FatalError("error restarting %s service: %v", constants.UPDATER_SERV_NAME, err)
	}
	h.Info("UTMStackUpdater restarted correctly")

	err = utils.RemoveLock(filepath.Join(path, "locks", constants.SERV_LOCK))
	if err != nil {
		h.FatalError("error removing %s: %v", constants.SERV_LOCK, err)
	}
	h.Info("UTMStackUpdater updated correctly")
}
