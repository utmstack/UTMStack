package serv

import (
	"fmt"

	"github.com/utmstack/UTMStack/collector-installer/config"

	"github.com/kardianos/service"
)

func InstallService(cnf config.ServiceConfig) error {
	svcConfig := getConfigServ(cnf.Name, cnf.DisplayName, cnf.Description)
	prg := &program{
		cmdRun:  cnf.CMDRun,
		cmdArgs: cnf.CMDArgs,
		path:    cnf.CMDPath,
	}

	newService, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("error creating new service: %v", err)
	}

	err = newService.Install()
	if err != nil {
		return fmt.Errorf("error installing new service: %v", err)
	}

	err = newService.Start()
	if err != nil {
		return fmt.Errorf("error starting new service: %v", err)
	}

	return nil
}
