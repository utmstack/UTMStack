package services

import (
	"fmt"

	"github.com/utmstack/UTMStack/agent/runner/depend"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func ConfigureServices(bins depend.ServicesBin, ip, utmKey, skip, config string) error {
	err := execBin(bins.AgentServiceBin, config, ip, utmKey, skip)
	if err != nil {
		return fmt.Errorf("error %sing UTMStackAgent service: %v", config, err)
	}

	if config == "install" {
		err = execBin(bins.RedlineServiceBin, config)
		if err != nil {
			return fmt.Errorf("error %sing UTMStackRedline service: %v", config, err)
		}

		err = execBin(bins.UpdaterServiceBin, config)
		if err != nil {
			return fmt.Errorf("error %sing UTMStackUpdater service: %v", config, err)
		}
	} else if config == "uninstall" {
		err = utils.UninstallService("UTMStackRedline")
		if err != nil {
			return fmt.Errorf("error %sing UTMStackRedline service: %v", config, err)
		}
		err = utils.UninstallService("UTMStackUpdater")
		if err != nil {
			return fmt.Errorf("error %sing UTMStackUpdater service: %v", config, err)
		}
	}

	return err
}
