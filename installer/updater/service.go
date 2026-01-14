package updater

import (
	"time"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/setup"
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
	go MonitorConnection(config.GetCMServer(), 30*time.Second, 3, &config.ConnectedToInternet)
	time.Sleep(5 * time.Second)

	// Check for pending update and apply it
	pendingUpdate, err := GetPendingUpdate()
	if err != nil {
		config.Logger().ErrorF("error getting pending update: %v", err)
	}

	if pendingUpdate != nil {
		config.Logger().Info("Applying pending update: %s-%s", pendingUpdate.Version, pendingUpdate.Edition)

		// Apply setup with the pending version
		if _, err := setup.Apply(pendingUpdate.Version); err != nil {
			config.Logger().ErrorF("error applying setup for version %s: %v", pendingUpdate.Version, err)
		} else {
			config.Logger().Info("Successfully applied update %s", pendingUpdate.Version)

			// Mark as sent in CM after successful apply
			if pendingUpdate.ID != "offline" {
				client := GetUpdaterClient()
				if err := client.MarkUpdateSent(pendingUpdate.ID); err != nil {
					config.Logger().ErrorF("error marking update %s as sent: %v", pendingUpdate.ID, err)
				}
			}
		}

		// Clear pending update
		if err := ClearPendingUpdate(); err != nil {
			config.Logger().ErrorF("error clearing pending update: %v", err)
		}
	}

	client := GetUpdaterClient()
	go UpdateWindowConfig()
	go client.UpdateProcess()
	go SyncSystemLogs()
	client.LicenseProcess()
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
