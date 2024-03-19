package serv

import (
	"github.com/kardianos/service"
	"github.com/threatwinds/logger"
)

func RunService(h *logger.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		h.Fatal("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		h.Fatal("error running new service: %v", err)
	}
}
