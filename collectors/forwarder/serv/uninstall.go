package serv

import (
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"github.com/utmstack/UTMStack/shared/svc"
)

func UninstallService() {
	// Stop is best-effort — the service may already be stopped
	if err := svc.Stop(config.ServiceName); err != nil {
		utils.Logger.ErrorF("could not stop %s (may already be stopped): %v", config.ServiceName, err)
	}
	if err := utils.UninstallService(config.ServiceName); err != nil {
		utils.Logger.Fatal("error uninstalling %s: %v", config.ServiceName, err)
	}
}
