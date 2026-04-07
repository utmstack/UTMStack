package service

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/agent/updater/updates"
	"github.com/utmstack/UTMStack/shared/logger"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		logger.Error("error getting config: %v", err)
		return
	}

	updates.UpdateDependencies(cnf)
}

func RunService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		logger.Fatal("error creating service: %v", err)
	}

	err = newService.Run()
	if err != nil {
		logger.Fatal("error running service: %v", err)
	}
}
