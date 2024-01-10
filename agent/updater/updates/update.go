package updates

import (
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
)

func UpdateServices(cnf configuration.Config, utmLogger *holmes.Logger) {
	utmServices := GetUTMServicesInstance()

	env, err := configuration.ReadEnv()
	if err != nil {
		utmLogger.FatalError("error reading environment configuration: %v", err)
	}

	utmLogger.Info("enviroment: %v", env)

	for {
		time.Sleep(configuration.CHECK_EVERY)

		err = utmServices.UpdateCurrentMasterVersion(cnf)
		if err != nil {
			utmLogger.Error("error updating current master version: %v", err)
			continue
		}

		err = utmServices.UpdateCurrentVersions()
		if err != nil {
			utmLogger.Error("error updating current versions: %v", err)
			continue
		}

		err = utmServices.UpdateLatestVersions(env, utmLogger)
		if err != nil {
			utmLogger.Error("error updating latest versions: %v", err)
			continue
		}

		err = utmServices.CheckUpdates(env, utmLogger)
		if err != nil {
			utmLogger.Error("error updating services: %v", err)
			continue
		}

	}
}
