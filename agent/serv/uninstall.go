package serv

import (
	"github.com/utmstack/UTMStack/agent/utils"
)

func UninstallService() {
	err := utils.StopService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackAgent: %v", err)
	}
	err = utils.StopService("UTMStackUpdater")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackUpdater: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackUpdater")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackUpdater: %v", err)
	}
}
