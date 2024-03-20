package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/redline/configuration"
	"github.com/utmstack/UTMStack/agent/redline/serv"
	"github.com/utmstack/UTMStack/agent/redline/utils"
)

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		log.Fatalf("Failed to get current path: %v", err)
	}

	// Configuring log saving
	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "run":
			serv.RunService(h)
		case "install":
			h.Info("Installing UTMStack Redline service...")
			serv.InstallService(h)
			h.Info("UTMStack Redline service installed correctly")

		case "uninstall":
			h.Info("Uninstalling UTMStack Redline service...")

			serv.UninstallService(h)
			h.Info("UTMStack Redline service uninstalled correctly")
			os.Exit(0)
		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService(h)
	}
}
