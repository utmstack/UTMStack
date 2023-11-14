package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/constants"
	"github.com/utmstack/UTMStack/agent/updater/serv"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

var h = holmes.New("debug", constants.SERV_NAME)

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	// Configuring log saving
	var logger = utils.CreateLogger(filepath.Join(path, "logs", constants.SERV_LOG))
	defer logger.Close()
	log.SetOutput(logger)

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "run":
			serv.RunService(h)
		case "install":
			h.Info("Installing UTMStack Updater service...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled(constants.SERV_NAME); err != nil {
				h.FatalError("error checking %s service: %v", constants.SERV_NAME, err)
			} else if isInstalled {
				h.FatalError("%s is already installed", constants.SERV_NAME)
			}

			serv.InstallService(h)
			h.Info("UTMStack Updater service installed correctly.")
			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			h.Info("Uninstalling UTMStack Updater service...")
			serv.UninstallService(h)
			h.Info("UTMStack Updater service uninstalled correctly.")
			time.Sleep(5 * time.Second)
			os.Exit(0)

		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService(h)
	}
}
