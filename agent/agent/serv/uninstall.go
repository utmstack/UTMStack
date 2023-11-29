package serv

import (
	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func UninstallService(h *holmes.Logger) {
	// Uninstall service
	err := utils.StopService("UTMStackAgent")
	if err != nil {
		h.FatalError("error stopping UTMStackAgent: %v", err)
	}
	err = utils.UninstallService("UTMStackAgent")
	if err != nil {
		h.FatalError("error uninstalling UTMStackAgent: %v", err)
	}
}
