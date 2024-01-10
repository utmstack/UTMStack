package services

import (
	"fmt"
	"strings"

	"github.com/utmstack/UTMStack/agent/runner/depend"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

func ConfigureServices(bins depend.ServicesBin, ip, utmKey, skip, config string) error {
	err := execBin(bins.AgentServiceBin, config, ip, utmKey, skip)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return fmt.Errorf("error %sing UTMStackAgent service: Check the file /logs/utmstack_agent.log for more details", config)
		}
		return fmt.Errorf("error %sing UTMStackAgent service: %v", config, err)
	}

	if config == "install" {
		err = execBin(bins.RedlineServiceBin, config)
		if err != nil {
			if strings.Contains(err.Error(), "exit status 1") {
				return fmt.Errorf("error %sing UTMStackRedline service: Check the file /logs/utmstack_redline.log for more details", config)
			}
			return fmt.Errorf("error %sing UTMStackRedline service: %v", config, err)
		}

		err = execBin(bins.UpdaterServiceBin, config)
		if err != nil {
			if strings.Contains(err.Error(), "exit status 1") {
				return fmt.Errorf("error %sing UTMStackUpdater service: Check the file /logs/utmstack_updater.log for more details", config)
			}
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
