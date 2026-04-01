package updater

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/docker"
	"github.com/utmstack/UTMStack/installer/setup"
	"github.com/utmstack/UTMStack/installer/utils"
)

func GetConfigServ() *service.Config {
	svcConfig := &service.Config{
		Name:        "UTMStackComponentsUpdater",
		DisplayName: "UTMStack Components Updater",
		Description: "UTMStack Components Updater",
		Executable:  config.InstallerBinPath,
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
	// Migrate service to standard path if needed
	if migrated := migrateServiceIfNeeded(); migrated {
		return // Exit, new service will start from standard path
	}

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
		if _, err := setup.Apply(pendingUpdate.Version, true); err != nil {
			config.Logger().ErrorF("error applying setup for version %s: %v", pendingUpdate.Version, err)
		} else {
			config.Logger().Info("Successfully applied update %s", pendingUpdate.Version)

			// Close ports after update (same as initial installation)
			config.Logger().Info("Running post-installation scripts to secure ports...")
			if err := docker.PostInstallation(); err != nil {
				config.Logger().ErrorF("error running post-installation: %v", err)
				config.Logger().ErrorF("CRITICAL: Post-installation failed. Stopping Docker to prevent security risk.")
				// Stop Docker to prevent leaving ports open
				if stopErr := utils.StopService("docker"); stopErr != nil {
					config.Logger().ErrorF("Failed to stop Docker service: %v", stopErr)
				} else {
					config.Logger().ErrorF("Docker service has been stopped. Manual intervention required.")
				}
				// Stop the updater service to prevent further processing
				return
			}

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
	// Copy current binary to standard location
	if err := copyInstallerToStandardPath(); err != nil {
		config.Logger().Fatal("error copying installer to standard path: %v", err)
	}

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

func copyInstallerToStandardPath() error {
	currentExec, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting current executable path: %v", err)
	}

	// If already at standard path, skip copy
	if currentExec == config.InstallerBinPath {
		return nil
	}

	srcFile, err := os.Open(currentExec)
	if err != nil {
		return fmt.Errorf("error opening source binary: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(config.InstallerBinPath)
	if err != nil {
		return fmt.Errorf("error creating destination binary: %v", err)
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("error copying binary: %v", err)
	}

	if err = os.Chmod(config.InstallerBinPath, 0755); err != nil {
		return fmt.Errorf("error setting permissions on binary: %v", err)
	}

	return nil
}

func migrateServiceIfNeeded() bool {
	currentExec, err := os.Executable()
	if err != nil {
		config.Logger().ErrorF("error getting current executable path: %v", err)
		return false
	}

	// Already running from standard path, no migration needed
	if currentExec == config.InstallerBinPath {
		return false
	}

	config.Logger().Info("Migrating service to standard path: %s", config.InstallerBinPath)

	// Copy current binary to standard location
	if err := copyInstallerToStandardPath(); err != nil {
		config.Logger().ErrorF("error copying installer during migration: %v", err)
		return false
	}

	// Uninstall old service
	serviceName := GetConfigServ().Name
	if err := utils.UninstallService(serviceName); err != nil {
		config.Logger().ErrorF("error uninstalling old service during migration: %v", err)
		return false
	}

	// Install new service pointing to standard path
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		config.Logger().ErrorF("error creating new service during migration: %v", err)
		return false
	}

	if err := newService.Install(); err != nil {
		config.Logger().ErrorF("error installing new service during migration: %v", err)
		return false
	}

	if err := newService.Start(); err != nil {
		config.Logger().ErrorF("error starting new service during migration: %v", err)
		return false
	}

	config.Logger().Info("Service migrated successfully to %s", config.InstallerBinPath)
	return true
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
