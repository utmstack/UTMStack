package serv

import (
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

// UninstallService uninstalls the service in the system
func UninstallService(utmLogger *logger.Logger) {
	// Uninstall service
	err := utils.StopService(configuration.SERV_NAME)
	if err != nil {
		utmLogger.Fatal("error stopping %s: %v", configuration.SERV_NAME, err)
	}
	err = utils.UninstallService(configuration.SERV_NAME)
	if err != nil {
		utmLogger.Fatal("error uninstalling %s: %v", configuration.SERV_NAME, err)
	}
}
