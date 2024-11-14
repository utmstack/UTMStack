package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/service/utils"
)

func RunService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		utils.Logger.Fatal("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		utils.Logger.Fatal("error running new service: %v", err)
	}
}
