package serv

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func CleanOldServices(cnf *config.Config) {
	oldVersion := false

	isUpdaterINstalled, err := utils.CheckIfServiceIsInstalled("UTMStackUpdater")
	if err != nil {
		utils.Logger.ErrorF("Error checking if service is installed: %v", err)
	}

	if isUpdaterINstalled {
		oldVersion = true
		err = utils.StopService("UTMStackUpdater")
		if err != nil {
			utils.Logger.ErrorF("Error stopping service: %v", err)
		}

		err = utils.UninstallService("UTMStackUpdater")
		if err != nil {
			utils.Logger.ErrorF("Error uninstalling service: %v", err)
		}
	}

	isRedlineInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackRedline")
	if err != nil {
		utils.Logger.ErrorF("Error checking if service is installed: %v", err)
	}

	if isRedlineInstalled {
		oldVersion = true
		err = utils.StopService("UTMStackRedline")
		if err != nil {
			utils.Logger.ErrorF("Error stopping service: %v", err)
		}

		err = utils.UninstallService("UTMStackRedline")
		if err != nil {
			utils.Logger.ErrorF("Error uninstalling service: %v", err)
		}
	}

	if oldVersion {
		headers := map[string]string{
			"key":  cnf.AgentKey,
			"id":   fmt.Sprintf("%v", cnf.AgentID),
			"type": "agent",
		}

		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, fmt.Sprintf(config.UpdaterSelf, "")), headers, fmt.Sprintf(config.UpdaterSelf, "_new"), utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
			utils.Logger.ErrorF("error downloading updater: %v", err)
			return
		}

		os.Remove(fmt.Sprintf(config.UpdaterSelf, ""))
		os.Rename(fmt.Sprintf(config.UpdaterSelf, "_new"), fmt.Sprintf(config.UpdaterSelf, ""))
	}
}
