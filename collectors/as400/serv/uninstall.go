package serv

import (
	"github.com/utmstack/UTMStack/collectors/as400/utils"
)

func UninstallService() {
	err := utils.StopService("UTMStackAS400Collector")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackAS400Collector: %v", err)
	}
	err = utils.UninstallService("UTMStackAS400Collector")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackAS400Collector: %v", err)
	}
}
