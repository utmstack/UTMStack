package main

import (
	"fmt"
	"os"

	pb "github.com/utmstack/UTMStack/collectors/as400/agent"
	"github.com/utmstack/UTMStack/collectors/as400/config"
	"github.com/utmstack/UTMStack/collectors/as400/serv"
	"github.com/utmstack/UTMStack/collectors/as400/updates"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
)

func main() {
	utils.InitLogger(config.ServiceLogFile)

	if len(os.Args) > 1 {
		arg := os.Args[1]

		isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackAS400Collector")
		if err != nil {
			fmt.Println("Error checking if service is installed: ", err)
			os.Exit(1)
		}
		if arg != "install" && !isInstalled {
			fmt.Println("UTMStackAS400Collector service is not installed")
			os.Exit(1)
		} else if arg == "install" && isInstalled {
			fmt.Println("UTMStackAS400Collector service is already installed")
			os.Exit(1)
		}

		switch arg {
		case "run":
			serv.RunService()
		case "install":
			utils.PrintBanner()
			fmt.Println("Installing UTMStackAS400Collector service ...")

			cnf, utmKey := config.GetInitialConfig()

			fmt.Print("Checking server connection ... ")
			if err := utils.ArePortsReachable(cnf.Server, config.AgentManagerPort, config.LogAuthProxyPort, config.DependenciesPort); err != nil {
				fmt.Println("\nError trying to connect to server: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Downloading dependencies ... ")
			if err := updates.DownloadVersion(cnf.Server, cnf.SkipCertValidation); err != nil {
				fmt.Println("\nError downloading version: ", err)
				os.Exit(1)
			}
			if err := updates.DownloadJar(cnf.Server, cnf.SkipCertValidation); err != nil {
				fmt.Println("\nError downloading jar: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Installing Java ... ")
			if err := utils.InstallJava(); err != nil {
				fmt.Println("\nError installing java: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print("Configuring collector ... ")
			err = pb.RegisterCollector(cnf, utmKey)
			if err != nil {
				fmt.Println("\nError registering collector: ", err)
				os.Exit(1)
			}
			if err = config.SaveConfig(cnf); err != nil {
				fmt.Println("\nError saving config: ", err)
				os.Exit(1)
			}
			if err = config.CreateManualConfigTemplate(); err != nil {
				fmt.Println("\nError creating config template: ", err)
				os.Exit(1)
			}
			fmt.Println("[OK]")

			fmt.Print(("Creating service ... "))
			serv.InstallService()
			fmt.Println("[OK]")

			fmt.Println("UTMStackAS400Collector service installed correctly")

		case "uninstall":
			fmt.Println("Uninstalling UTMStackAS400Collector service ...")

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				fmt.Println("Error getting config: ", err)
				os.Exit(1)
			}
			if err = pb.DeleteAgent(cnf); err != nil {
				utils.Logger.ErrorF("error deleting collector: %v", err)
			}

			os.Remove(config.ConfigurationFile)

			serv.UninstallService()

			fmt.Println("Uninstalling java")
			if err := utils.UninstallJava(); err != nil {
				utils.Logger.ErrorF("error unistalling java: %v", err)
			}

			fmt.Println("[OK]")
			fmt.Println("UTMStackAS400Collector service uninstalled correctly")
			os.Exit(1)
		case "help":
			Help()
		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService()
	}
}

func Help() {
	fmt.Println("### UTMStack Collector ###")
	fmt.Println("Usage:")
	fmt.Println("  To run the service:                     ./utmstack_as400_collector run")
	fmt.Println("  To install the service:                 ./utmstack_as400_collector install")
	fmt.Println("  To uninstall the service:               ./utmstack_as400_collector uninstall")
	fmt.Println("  To debug UTMStack installation:         ./utmstack_as400_collector debug-utmstack")
	fmt.Println("  For help (this message):                ./utmstack_as400_collector help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  run                      Run the UTMStackAS400Collector service")
	fmt.Println("  install                  Install the UTMStackAS400Collector service")
	fmt.Println("  uninstall                Uninstall the UTMStackAS400Collector service")
	fmt.Println("  debug-utmstack           Debug UTMStack installation validation")
	fmt.Println("  help                     Display this help message")
	fmt.Println()
	fmt.Println("Requirements:")
	fmt.Println("  - UTMStack must be installed on this system")
	fmt.Println("  - File /utmstack.yaml must exist in root directory")
	fmt.Println("  - Directory /utmstack/ must exist")
	fmt.Println()
	fmt.Println("Note:")
	fmt.Println("  - Make sure to run commands with appropriate permissions.")
	fmt.Println("  - All commands require administrative privileges.")
	fmt.Println("  - For detailed logs, check the service log file.")
	fmt.Println()
	os.Exit(0)
}
