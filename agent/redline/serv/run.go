package serv

import (
	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
)

func RunService(h *holmes.Logger) {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		h.FatalError("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		h.FatalError("error running new service: %v", err)
	}
}
