package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/runner/checkversion"
	"github.com/utmstack/UTMStack/agent/runner/configuration"
	"github.com/utmstack/UTMStack/agent/runner/depend"
	"github.com/utmstack/UTMStack/agent/runner/services"
	"github.com/utmstack/UTMStack/agent/runner/utils"
)

var h = holmes.New("debug", "UTMStackAgentInstaller")

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	servBins := depend.ServicesBin{}
	switch runtime.GOOS {
	case "windows":
		servBins.AgentServiceBin = "utmstack_agent_service.exe"
		servBins.UpdaterServiceBin = "utmstack_updater_service.exe"
		servBins.RedlineServiceBin = "utmstack_redline_service.exe"
	case "linux":
		servBins.AgentServiceBin = "utmstack_agent_service"
		servBins.UpdaterServiceBin = "utmstack_updater_service"
		servBins.RedlineServiceBin = "utmstack_redline_service"
	}

	// Configuring log saving
	var logger = utils.CreateLogger(filepath.Join(path, "logs", "utmstack_agent_installer.log"))
	defer logger.Close()
	log.SetOutput(logger)

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "install":
			ip, utmKey, skip := os.Args[2], os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				fmt.Println("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.")
				h.FatalError("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.")
			}

			fmt.Println("Installing UTMStack Agent services...")
			h.Info("Installing UTMStack Agent services...")

			if !utils.IsPortOpen(ip, configuration.AgentManagerPort) || !utils.IsPortOpen(ip, configuration.LogAuthProxyPort) {
				fmt.Printf("Error installing the UTMStack Agent: one or more of the requiered ports are closed. Please open ports 9000 and 50051.")
				h.FatalError("Error installing the UTMStack Agent: one or more of the requiered ports are closed. Please open ports 9000 and 50051.")
			}

			err = checkversion.CleanOldVersions(h)
			if err != nil {
				fmt.Printf("error cleaning old versions: %v", err)
				h.FatalError("error cleaning old versions: %v", err)
			}

			// Download dependencies
			err := depend.DownloadDependencies(servBins, ip)
			if err != nil {
				fmt.Printf("error downloading dependencies: %v", err)
				h.FatalError("error downloading dependencies: %v", err)
			}

			// Install Services
			err = services.ConfigureServices(servBins, ip, utmKey, skip, "install")
			if err != nil {
				fmt.Printf("error installing UTMStack services: %v", err)
				h.FatalError("error installing UTMStack services: %v", err)
			}

			h.Info("UTMStack Agent services installed correctly.")
			fmt.Println("UTMStack Agent services installed correctly.")
			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			fmt.Println("Uninstalling UTMStack services...")

			// Uninstall Services
			if isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent"); err != nil {
				fmt.Printf("error checking UTMStackAgent service: %v", err)
				h.FatalError("error checking UTMStackAgent service: %v", err)
			} else if isInstalled {
				err = services.ConfigureServices(servBins, "", "", "", "uninstall")
				if err != nil {
					fmt.Printf("error uninstalling UTMStack services: %v", err)
					h.FatalError("error uninstalling UTMStack services: %v", err)
				}

				fmt.Println("UTMStack services uninstalled correctly.")
				time.Sleep(5 * time.Second)
				os.Exit(0)

			} else {
				fmt.Println("UTMStackAgent not installed")
				h.FatalError("UTMStackAgent not installed")
			}

		default:
			fmt.Println("unknown option")
		}
	}
}
