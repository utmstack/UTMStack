package serv

import (
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

func CleanOldServices(cnf *config.Config) {
	isRedlineInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackRedline")
	if err != nil {
		utils.Logger.LogF(100, "error checking if UTMStackRedline service is installed: %v", err)
		return
	}

	if isRedlineInstalled {
		utils.Logger.Info("old UTMStackRedline service found, removing...")

		err = utils.StopService("UTMStackRedline")
		if err != nil {
			utils.Logger.LogF(100, "error stopping UTMStackRedline service: %v", err)
		}

		err = utils.UninstallService("UTMStackRedline")
		if err != nil {
			utils.Logger.LogF(100, "error uninstalling UTMStackRedline service: %v", err)
		} else {
			utils.Logger.Info("UTMStackRedline service removed successfully")
		}
	}
}
