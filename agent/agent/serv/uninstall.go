package serv

import (
	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func UninstallService(h *logger.Logger) {
	// Uninstall service
	err := utils.StopService("UTMStackAgent")
	if err != nil {
		h.Fatal("error stopping UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		h.Fatal("error uninstalling UTMStackAgent: %v", err)
	}
}
