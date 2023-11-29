package serv

import (
	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/constants"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

func UninstallService(h *holmes.Logger) {
	// Uninstall service
	err := utils.StopService(constants.SERV_NAME)
	if err != nil {
		h.FatalError("error stopping %s: %v", constants.SERV_NAME, err)
	}
	err = utils.UninstallService(constants.SERV_NAME)
	if err != nil {
		h.FatalError("error uninstalling %s: %v", constants.SERV_NAME, err)
	}
}
