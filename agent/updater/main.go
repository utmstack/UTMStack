package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/updater/config"
	"github.com/utmstack/UTMStack/agent/updater/service"
	"github.com/utmstack/UTMStack/shared/fs"
	"github.com/utmstack/UTMStack/shared/logger"
)

func main() {
	basePath := fs.GetExecutablePath()
	logger.Init(filepath.Join(basePath, "logs", config.SERV_LOG), logger.LevelInfo)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "install":
			fmt.Println("Installing UTMStack Updater service...")

			fmt.Print("Creating service ... ")
			service.InstallService()
			fmt.Println("[OK]")

			fmt.Println("UTMStackUpdater service installed correctly")
			return
		case "uninstall":
			fmt.Println("Uninstalling UTMStack Updater service...")
			service.UninstallService()
			fmt.Println("Service uninstalled successfully")
			return
		case "start":
			fmt.Println("Starting UTMStack Updater service...")
			return
		case "stop":
			fmt.Println("Stopping UTMStack Updater service...")
			return
		}
	}

	service.RunService()
}
