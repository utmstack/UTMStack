package serv

import (
	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
)

// RunService runs the service in the system
func RunService(utmLogger *holmes.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		utmLogger.FatalError("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		utmLogger.FatalError("error running new service: %v", err)
	}
}
