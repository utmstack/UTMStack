package serv

import (
	"github.com/kardianos/service"
	"github.com/threatwinds/logger"
)

func InstallService(h *logger.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		h.Fatal("error creating new service: %v", err)
	}
	err = newService.Install()
	if err != nil {
		h.Fatal("error installing new service: %v", err)
	}

	// Start the service after installing it
	err = newService.Start()
	if err != nil {
		h.Fatal("error starting new service: %v", err)
	}
}
