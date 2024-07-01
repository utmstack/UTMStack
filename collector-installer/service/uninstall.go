package serv

import (
	"fmt"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/utils"
)

func UninstallService(cnf config.ServiceConfig) error {
	utils.StopService(cnf.Name)
	err := utils.UninstallService(cnf.Name)
	if err != nil {
		return fmt.Errorf("error uninstalling service: %v", err)
	}
	return nil
}
