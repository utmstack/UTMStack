package updates

import (
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/constants"
)

func UpdateServices(cnf configuration.Config, h *holmes.Logger) {
	utmServices := GetUTMServicesInstance()

	for {
		time.Sleep(constants.CheckUpdatesEvery)

		masterVersion, err := getMasterVersion(cnf.Server, cnf.SkipCertValidation)
		if err != nil {
			h.Error("error getting master version: %v", err)
			continue
		}

		// save current versions
		err = utmServices.UpdateCurrentVersions()
		if err != nil {
			h.Error("error updating current versions: %v", err)
			continue
		}

		// download new versions and save
		err = utmServices.UpdateLatestVersions()
		if err != nil {
			h.Error("error updating latest versions: %v", err)
			continue
		}

		err = utmServices.CheckUpdates(masterVersion, h)
		if err != nil {
			h.Error("error updating services: %v", err)
			continue
		}

	}
}
