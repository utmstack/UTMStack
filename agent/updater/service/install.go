package service

import (
	"fmt"
	"os"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/shared/logger"
	"github.com/utmstack/UTMStack/shared/svc"
)

func InstallService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println("\nError creating new service: ", err)
		os.Exit(1)
	}
	err = newService.Install()
	if err != nil {
		fmt.Println("\nError installing new service: ", err)
		os.Exit(1)
	}

	if err := svc.EnsureRecovery(svcConfig.Name); err != nil {
		fmt.Printf("\nWarning configuring updater service recovery: %v\n", err)
	}

	err = newService.Start()
	if err != nil {
		fmt.Println("\nError starting new service: ", err)
		os.Exit(1)
	}
	logger.Info("updater service installed successfully")
}

func UninstallService() {
	svcConfig := GetConfigServ()
	prg := new(program)
	newService, err := service.New(prg, svcConfig)
	if err != nil {
		fmt.Println("\nError creating new service: ", err)
		os.Exit(1)
	}

	err = newService.Stop()
	if err != nil {
		fmt.Println("\nWarning stopping service: ", err)
	}

	err = newService.Uninstall()
	if err != nil {
		fmt.Println("\nError uninstalling service: ", err)
		os.Exit(1)
	}
}
