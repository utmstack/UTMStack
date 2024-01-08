package serv

import (
	"fmt"

	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
)

// InstallService installs the service in the system and starts it
func InstallService(utmLogger *holmes.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Printf("error creating new service: %v", err)
		utmLogger.FatalError("error creating new service: %v", err)
	}
	err = newService.Install()
	if err != nil {
		fmt.Printf("error installing new service: %v", err)
		utmLogger.FatalError("error installing new service: %v", err)
	}

	// Start the service after installing it
	err = newService.Start()
	if err != nil {
		fmt.Printf("error starting new service: %v", err)
		utmLogger.FatalError("error starting new service: %v", err)
	}
}
