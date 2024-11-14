package main

import (
	"os"
	"path/filepath"
	"time"

	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/database"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/models"
	"github.com/utmstack/UTMStack/agent/agent/modules"
	"github.com/utmstack/UTMStack/agent/agent/serv"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

func main() {
	path := utils.GetMyPath()
	utils.InitLogger(filepath.Join(path, "logs", config.SERV_LOG))

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "run":
			serv.RunService()
		case "install":
			utils.Logger.Info("Installing UTMStack Agent service ...")

			cnf, utmKey := config.GetInitialConfig()

			err := conn.EstablishConnectionsFistTime(cnf)
			if err != nil {
				utils.Logger.Fatal("error establishing connections: %v", err)
			}

			if err = pb.RegisterAgent(cnf, utmKey); err != nil {
				utils.Logger.Fatal("%v", err)
			}

			if err = config.SaveConfig(cnf); err != nil {
				utils.Logger.Fatal("error writing config file: %v", err)
			}

			if err = modules.ConfigureCollectorFirstTime(); err != nil {
				utils.Logger.Fatal("error configuring syslog server: %v", err)
			}

			if err = beats.InstallBeats(*cnf); err != nil {
				utils.Logger.Fatal("error installing beats: %v", err)
			}

			if err := logservice.SetDataRetention(""); err != nil {
				utils.Logger.Fatal("error trying to change retention: %v", err)
			}

			serv.InstallService()
			utils.Logger.Info("UTMStack Agent service installed correctly")

		case "enable-integration", "disable-integration":
			integration := os.Args[2]
			proto := os.Args[3]

			port, err := modules.ChangeIntegrationStatus(integration, proto, (arg == "enable-integration"))
			if err != nil {
				utils.Logger.ErrorF("error trying to %s: %v", arg, err)
				os.Exit(0)
			}
			utils.Logger.Info("%s %s done correctly in port %s %s", arg, integration, port, proto)
			time.Sleep(5 * time.Second)

		case "change-port":
			integration := os.Args[2]
			proto := os.Args[3]
			port := os.Args[4]

			old, err := modules.ChangePort(integration, proto, port)
			if err != nil {
				utils.Logger.ErrorF("error trying to change port: %v", err)
				os.Exit(0)
			}
			utils.Logger.Info("change port done correctly from %s to %s in %s for %s integration", old, port, proto, integration)
			time.Sleep(5 * time.Second)

		case "change-retention":
			retention := os.Args[2]

			if err := logservice.SetDataRetention(retention); err != nil {
				utils.Logger.Fatal("error trying to change retention: %v", err)
			}

			utils.Logger.Info("change retention done correctly to %s", retention)

		case "clean-logs":
			db := database.GetDB()
			datR, err := logservice.GetDataRetention()
			if err != nil {
				utils.Logger.ErrorF("error getting retention: %v", err)
				os.Exit(0)
			}
			_, err = db.DeleteOld(models.Log{}, datR)
			if err != nil {
				utils.Logger.ErrorF("error deleting old logs: %v", err)
				os.Exit(0)
			}

		case "uninstall":
			utils.Logger.Info("Uninstalling UTMStack Agent service...")

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				utils.Logger.Fatal("error getting config: %v", err)
			}

			err = conn.EstablishConnectionsFistTime(cnf)
			if err != nil {
				utils.Logger.Fatal("error establishing connections: %v", err)
			}

			if err = pb.DeleteAgent(cnf); err != nil {
				utils.Logger.ErrorF("error deleting agent: %v", err)
			}

			if err = beats.UninstallBeats(); err != nil {
				utils.Logger.Fatal("error uninstalling beats: %v", err)
			}
			os.Remove(filepath.Join(path, "config.yml"))

			serv.UninstallService()
			utils.Logger.Info("UTMStack Agent service uninstalled correctly")
			os.Exit(0)
		default:
			utils.Logger.Info("unknown option")
		}
	} else {
		serv.RunService()
	}
}
