package updates

import "github.com/utmstack/UTMStack/agent/service/utils"

func RemoveOldServices() error {
	if isInstalled, _ := utils.CheckIfServiceIsInstalled("UTMStackUpdater"); isInstalled {
		utils.StopService("UTMStackUpdater")
		utils.UninstallService("UTMStackUpdater")
	}

	if isInstalled, _ := utils.CheckIfServiceIsInstalled("UTMStackRedline"); isInstalled {
		utils.StopService("UTMStackRedline")
		utils.UninstallService("UTMStackRedline")
	}
	return nil
}
