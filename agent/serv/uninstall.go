package serv

import (
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/svc"
)

func UninstallService() {
	err := svc.Stop("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error stopping UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		utils.Logger.Fatal("error uninstalling UTMStackAgent: %v", err)
	}
}
