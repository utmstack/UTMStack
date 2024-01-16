package serv

import (
	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

// UninstallService uninstalls the service in the system
func UninstallService(utmLogger *holmes.Logger) {
	// Uninstall service
	err := utils.StopService(configuration.SERV_NAME)
	if err != nil {
		utmLogger.FatalError("error stopping %s: %v", configuration.SERV_NAME, err)
	}
	err = utils.UninstallService(configuration.SERV_NAME)
	if err != nil {
		utmLogger.FatalError("error uninstalling %s: %v", configuration.SERV_NAME, err)
	}
}
