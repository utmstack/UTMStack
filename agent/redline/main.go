package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/redline/constants"
	"github.com/utmstack/UTMStack/agent/redline/serv"
	"github.com/utmstack/UTMStack/agent/redline/utils"
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
