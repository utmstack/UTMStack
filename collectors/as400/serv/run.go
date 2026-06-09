package serv

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
)

func RunService() {
	svcConfig := GetConfigServ()
	p := new(program)

	newService, err := service.New(p, svcConfig)
	if err != nil {
		utils.Logger.Fatal("error creating new service: %v", err)
	}

	err = newService.Run()
	if err != nil {
		utils.Logger.Fatal("error running new service: %v", err)
	}
}
