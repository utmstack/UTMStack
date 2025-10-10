package serv

import (
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
)

func UninstallService() {
	err := utils.StopService("UTMStackCollector")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackCollector: %v", err)
	}
	err = utils.UninstallService("UTMStackCollector")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackCollector: %v", err)
	}
}
