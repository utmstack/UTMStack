package updater

import (
	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackComponentsUpdater",
		DisplayName: "UTMStack Components Updater",
		Description: "UTMStack Components Updater",
		Arguments:   []string{"--run"},
	}

	return svcConfig
}

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	GetUpdaterClient().UpdateProcess()
}

func InstallService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		config.Logger().Fatal("error creating new service: %v", err)
	}
	err = newService.Install()
	if err != nil {
		config.Logger().Fatal("error installing new service: %v", err)
	}

	err = newService.Start()
	if err != nil {
		config.Logger().Fatal("error starting new service: %v", err)
	}
}

func RunService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		config.Logger().Fatal("error creating new service: %v", err)
	}
	err = newService.Run()
	if err != nil {
		config.Logger().Fatal("error running new service: %v", err)
	}
}

func UninstallService() {
	serviceName := GetConfigServ().Name
	err := utils.StopService(serviceName)
	if err != nil {
		config.Logger().Fatal("error stopping %s: %v", serviceName, err)
	}
	err = utils.UninstallService(serviceName)
	if err != nil {
		config.Logger().Fatal("error uninstalling %s: %v", serviceName, err)
	}
}
