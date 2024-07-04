package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/modules"
	"github.com/utmstack/UTMStack/agent/agent/serv"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc/metadata"
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

			conn, err := conn.ConnectToServer(cnf, cnf.Server, config.AGENTMANAGERPORT)
			if err != nil {
				utils.Logger.Fatal("error connecting to Agent Manager: %v", err)
			}
			defer conn.Close()
			utils.Logger.Info("Connection to Agent Manager successful!!!")

			if err = pb.RegisterAgent(conn, cnf, utmKey); err != nil {
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

			serv.InstallService()
			utils.Logger.Info("UTMStack Agent service installed correctly")

		case "send-log":
			msg := os.Args[2]
			logp := logservice.GetLogProcessor()

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				os.Exit(0)
			}

			connLogServ, err := conn.ConnectToServer(cnf, cnf.Server, config.AUTHLOGSPORT)
			if err != nil {
				utils.Logger.ErrorF("error connecting to Log Auth Proxy: %v", err)
			}
			defer connLogServ.Close()

			logClient := logservice.NewLogServiceClient(connLogServ)
			ctxLog, cancelLog := context.WithCancel(context.Background())
			defer cancelLog()
			ctxLog = metadata.AppendToOutgoingContext(ctxLog, "agent-key", cnf.AgentKey)

			err = logp.ProcessLogsWithHighPriority(msg, logClient, ctxLog, cnf)
			if err != nil {
				utils.Logger.ErrorF("error sending High Priority Log to Log Auth Proxy: %v", err)
			}
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

		case "uninstall":
			utils.Logger.Info("Uninstalling UTMStack Agent service...")

			cnf, err := config.GetCurrentConfig()
			if err != nil {
				utils.Logger.Fatal("error getting config: %v", err)
			}

			conn, err := conn.ConnectToServer(cnf, cnf.Server, config.AGENTMANAGERPORT)
			if err != nil {
				utils.Logger.ErrorF("error connecting to Agent Manager: %v", err)
			} else {
				utils.Logger.Info("Connection to Agent Manager successful!!!")

				if err = pb.DeleteAgent(conn, cnf); err != nil {
					utils.Logger.ErrorF("error deleting agent: %v", err)
				}
			}
			defer conn.Close()

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
