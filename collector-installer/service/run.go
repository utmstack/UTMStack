package serv

import (
	"fmt"

	"github.com/utmstack/UTMStack/collector-installer/config"
	"github.com/utmstack/UTMStack/collector-installer/utils"

	"github.com/kardianos/service"
)

func RunService(cnf config.ServiceConfig, logger *utils.BeautyLogger) error {
	svcConfig := getConfigServ(cnf.Name, cnf.DisplayName, cnf.Description)
	prg := &program{
		cmdRun:  cnf.CMDRun,
		cmdArgs: cnf.CMDArgs,
		path:    cnf.CMDPath,
		logger:  logger,
	}
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		return fmt.Errorf("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		return fmt.Errorf("error running new service: %v", err)
	}
	return nil
}
