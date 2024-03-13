package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/serv"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

func main() {
	path, err := utils.GetMyPath()
	if err != nil {
		log.Fatalf("failed to get current path: %v", err)
	}

	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "run":
			serv.RunService(h)
		case "install":
			h.Info("Installing UTMStack Updater service...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled(configuration.SERV_NAME); err != nil {
				h.Fatal("error checking %s service: %v", configuration.SERV_NAME, err)
			} else if isInstalled {
				h.Fatal("%s is already installed", configuration.SERV_NAME)
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
