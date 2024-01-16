package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/serv"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

var utmLogger = holmes.New("debug", configuration.SERV_NAME)

func main() {
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("failed to get current path: %v", err)
		utmLogger.FatalError("failed to get current path: %v", err)
	}

	var logger = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))
	defer logger.Close()
	log.SetOutput(logger)

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "run":
			serv.RunService(utmLogger)
		case "install":
			utmLogger.Info("Installing UTMStack Updater service...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled(configuration.SERV_NAME); err != nil {
				utmLogger.FatalError("error checking %s service: %v", configuration.SERV_NAME, err)
			} else if isInstalled {
				utmLogger.FatalError("%s is already installed", configuration.SERV_NAME)
			}

			serv.InstallService(utmLogger)
			utmLogger.Info("UTMStack Updater service installed correctly.")
			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			utmLogger.Info("Uninstalling UTMStack Updater service...")
			serv.UninstallService(utmLogger)
			utmLogger.Info("UTMStack Updater service uninstalled correctly.")
			time.Sleep(5 * time.Second)
			os.Exit(0)

		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService(utmLogger)
	}
}
