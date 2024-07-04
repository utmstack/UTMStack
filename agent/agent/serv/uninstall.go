package serv

import (
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func UninstallService() {
	err := utils.StopService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackAgent: %v", err)
	}
}
