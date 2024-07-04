package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func InstallService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		utils.Logger.Fatal("error creating new service: %v", err)
	}
	err = newService.Install()
	if err != nil {
		utils.Logger.Fatal("error installing new service: %v", err)
	}

	err = newService.Start()
	if err != nil {
		utils.Logger.Fatal("error starting new service: %v", err)
	}
}
