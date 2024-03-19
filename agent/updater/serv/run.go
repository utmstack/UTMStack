package serv

import (
	"github.com/kardianos/service"
	"github.com/threatwinds/logger"
)

// RunService runs the service in the system
func RunService(utmLogger *logger.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		utmLogger.Fatal("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		utmLogger.Fatal("error running new service: %v", err)
	}
}
