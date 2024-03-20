package main

import (
	"log"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/self/configuration"
	"github.com/utmstack/UTMStack/agent/self/update"
	"github.com/utmstack/UTMStack/agent/self/utils"
)

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		log.Fatalf("Failed to get current path: %v", err)
	}

	// Configuring log saving
	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))

	// Save data from versions.json
	allVersions := update.Version{}
	err = utils.ReadJson(filepath.Join(path, "versions.json"), &allVersions)
	if err != nil {
		h.Fatal("error reading current versions.json: %v", err)
	}

	// Select environment
	env, err := configuration.ReadEnv()
	if err != nil {
		h.Fatal("Failed to get current path: %v", err)
	}

	h.Info("Updating UTMStackUpdater...")

	if isRunning, err := utils.CheckIfServiceIsActive(configuration.UPDATER_SERV_NAME); err != nil {
		h.Fatal("error checking %s service: %v", configuration.UPDATER_SERV_NAME, err)
	} else if isRunning {
		err = utils.StopService(configuration.UPDATER_SERV_NAME)
		if err != nil {
			h.Fatal("error stopping %s service: %v", configuration.UPDATER_SERV_NAME, err)
		}
		h.Info("UTMStackUpdater stopped correctly")
	}

	err = update.UpdateUpdaterService(allVersions.UpdaterVersion, env.Branch)
	if err != nil {
		h.Fatal("error downloading new %s service: %v", configuration.UPDATER_SERV_NAME, err)
	}
	h.Info("New UTMStackUpdater downloaded correctly")

	err = utils.RestartService(configuration.UPDATER_SERV_NAME)
	if err != nil {
		h.Fatal("error restarting %s service: %v", configuration.UPDATER_SERV_NAME, err)
	}
	h.Info("UTMStackUpdater restarted correctly")

	err = utils.RemoveLock(filepath.Join(path, "locks", configuration.SERV_LOCK))
	if err != nil {
		h.Fatal("error removing %s: %v", configuration.SERV_LOCK, err)
	}
	h.Info("UTMStackUpdater updated correctly")
}
