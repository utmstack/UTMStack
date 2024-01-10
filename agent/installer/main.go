package main

import (
	"log"
	"os"
	"path/filepath"
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
	beautyLogger := utils.GetBeautyLogger()
	beautyLogger.PrintBanner()

	path, err := utils.GetMyPath()
	if err != nil {
		beautyLogger.WriteError("failed to get current path", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	servBins := depend.GetServicesBins()

	var logger = utils.CreateLogger(filepath.Join(path, "logs", "utmstack_agent_installer.log"))
	defer logger.Close()
	log.SetOutput(logger)

	if len(os.Args) > 1 {
		mode := os.Args[1]
		switch mode {
		case "install":
			ip, utmKey, skip := os.Args[2], os.Args[3], os.Args[4]

			if strings.Count(utmKey, "*") == len(utmKey) {
				beautyLogger.WriteError("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.", nil)
				h.FatalError("The connection key provided is incorrect. Please make sure you use the 'copy' icon from the integrations section to get the value of the masked key value.")
			}

			beautyLogger.WriteSimpleMessage("Installing UTMStack Agent...")
			if !utils.IsPortOpen(ip, configuration.AgentManagerPort) || !utils.IsPortOpen(ip, configuration.LogAuthProxyPort) {
				beautyLogger.WriteError("one or more of the requiered ports are closed. Please open ports 9000 and 50051.", nil)
				h.FatalError("Error installing the UTMStack Agent: one or more of the requiered ports are closed. Please open ports 9000 and 50051.")
			}

			err := utils.CreatePathIfNotExist(filepath.Join(path, "locks"))
			if err != nil {
				beautyLogger.WriteError("error creating locks path", err)
				h.FatalError("error creating locks path: %v", err)
			}

			err = utils.SetLock(filepath.Join(path, "locks", "setup.lock"))
			if err != nil {
				beautyLogger.WriteError("error setting setup.lock", err)
				h.FatalError("error setting setup.lock: %v", err)
			}

			err = checkversion.CleanOldVersions(h)
			if err != nil {
				beautyLogger.WriteError("error cleaning old versions", err)
				h.FatalError("error cleaning old versions: %v", err)
			}

			beautyLogger.WriteSimpleMessage("Downloading UTMStack dependencies...")
			err = depend.DownloadDependencies(servBins, ip, skip)
			if err != nil {
				beautyLogger.WriteError("error downloading dependencies", err)
				h.FatalError("error downloading dependencies: %v", err)
			}
			beautyLogger.WriteSuccessfull("UTMStack dependencies downloaded correctly.")

			beautyLogger.WriteSimpleMessage("Installing services...")
			err = services.ConfigureServices(servBins, ip, utmKey, skip, "install")
			if err != nil {
				beautyLogger.WriteError("error installing UTMStack services", err)
				h.FatalError("error installing UTMStack services: %v", err)
			}

			err = utils.RemoveLock(filepath.Join(path, "locks", "setup.lock"))
			if err != nil {
				beautyLogger.WriteError("error removing setup.lock", err)
				h.FatalError("error removing setup.lock: %v", err)
			}

			beautyLogger.WriteSuccessfull("Services installed correctly")
			beautyLogger.WriteSuccessfull("UTMStack Agent installed correctly.")

			time.Sleep(5 * time.Second)
			os.Exit(0)

		case "uninstall":
			beautyLogger.WriteSimpleMessage("Uninstalling UTMStack Agent...")

			if isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAgent"); err != nil {
				beautyLogger.WriteError("error checking UTMStackAgent service", err)
				h.FatalError("error checking UTMStackAgent service: %v", err)
			} else if isInstalled {
				beautyLogger.WriteSimpleMessage("Uninstalling UTMStack services...")
				err = services.ConfigureServices(servBins, "", "", "", "uninstall")
				if err != nil {
					beautyLogger.WriteError("error uninstalling UTMStack services", err)
					h.FatalError("error uninstalling UTMStack services: %v", err)
				}

				beautyLogger.WriteSuccessfull("UTMStack services uninstalled correctly.")
				time.Sleep(5 * time.Second)
				os.Exit(0)

			} else {
				beautyLogger.WriteError("UTMStackAgent not installed", nil)
				h.FatalError("UTMStackAgent not installed")
			}

		default:
			beautyLogger.WriteError("unknown option", nil)
		}
	}
}
