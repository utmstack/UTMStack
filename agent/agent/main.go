package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/quantfall/holmes"
	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/serv"
	"github.com/utmstack/UTMStack/agent/agent/syslog"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc/metadata"
)

var h = holmes.New("debug", configuration.SERV_NAME)

func main() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		fmt.Printf("Failed to get current path: %v", err)
		h.FatalError("Failed to get current path: %v", err)
	}

	var logger = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))
	defer logger.Close()
	log.SetOutput(logger)

	if len(os.Args) > 1 {
		arg := os.Args[1]
		switch arg {
		case "run":
			serv.RunService(h)
		case "install":
			h.Info("Installing UTMStack Agent service...")

			certsPath := filepath.Join(path, "certs")
			err = utils.CreatePathIfNotExist(certsPath)
			if err != nil {
				fmt.Printf("error creating path: %s", err)
				h.FatalError("error creating path: %s", err)
			}

			err = utils.GenerateCerts(certsPath)
			if err != nil {
				fmt.Printf("error generating certificates: %v", err)
				h.FatalError("error generating certificates: %v", err)
			}

			cnf, utmKey := configuration.GetInitialConfig()

			conn, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
			if err != nil {
				fmt.Printf("error connecting to Agent Manager: %v", err)
				h.FatalError("error connecting to Agent Manager: %v", err)
			}
			defer conn.Close()
			h.Info("Connection to Agent Manager successful!!!")

			if err = pb.RegisterAgent(conn, cnf, utmKey, h); err != nil {
				h.FatalError("%v", err)
			}

			// Write config in config.yml
			if err = configuration.SaveConfig(cnf); err != nil {
				fmt.Printf("error writing config file: %v", err)
				h.FatalError("error writing config file: %v", err)
			}

			if err = syslog.ConfigureCollectorFirstTime(); err != nil {
				fmt.Printf("error configuring syslog server: %v", err)
				h.FatalError("error configuring syslog server: %v", err)
			}

			if err = beats.InstallBeats(*cnf, h); err != nil {
				fmt.Printf("error installing beats: %v", err)
				h.FatalError("error installing beats: %v", err)
			}

			serv.InstallService(h)
			h.Info("UTMStack Agent service installed correctly")

		case "send-log":
			msg := os.Args[2]
			logp := logservice.GetLogProcessor()

			cnf, err := configuration.GetCurrentConfig()
			if err != nil {
				os.Exit(0)
			}

			connLogServ, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AUTHLOGSPORT)
			if err != nil {
				fmt.Printf("error connecting to Log Auth Proxy: %v", err)
				h.Error("error connecting to Log Auth Proxy: %v", err)
			}
			defer connLogServ.Close()

			logClient := logservice.NewLogServiceClient(connLogServ)
			ctxLog, cancelLog := context.WithCancel(context.Background())
			defer cancelLog()
			ctxLog = metadata.AppendToOutgoingContext(ctxLog, "agent-key", cnf.AgentKey)

			err = logp.ProcessLogsWithHighPriority(msg, logClient, ctxLog, cnf, h)
			if err != nil {
				h.Error("error sending High Priority Log to Log Auth Proxy: %v", err)
			}
		case "uninstall":
			h.Info("Uninstalling UTMStack Agent service...")

			cnf, err := configuration.GetCurrentConfig()
			if err != nil {
				fmt.Printf("error getting config: %v", err)
				h.FatalError("error getting config: %v", err)
			}

			conn, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
			if err != nil {
				fmt.Printf("error connecting to Agent Manager: %v", err)
				h.Error("error connecting to Agent Manager: %v", err)
			} else {
				h.Info("Connection to Agent Manager successful!!!")
				fmt.Printf("Connection to Agent Manager successful!!!")

				if err = pb.DeleteAgent(conn, cnf, h); err != nil {
					fmt.Printf("error deleting agent: %v", err)
					h.Error("error deleting agent: %v", err)
				}
			}
			defer conn.Close()

			if err = beats.UninstallBeats(h); err != nil {
				fmt.Printf("error uninstalling beats: %v", err)
				h.FatalError("error uninstalling beats: %v", err)
			}
			os.Remove(filepath.Join(path, "config.yml"))

			serv.UninstallService(h)
			h.Info("UTMStack Agent service uninstalled correctly")
			os.Exit(0)
		default:
			fmt.Println("unknown option")
		}
	} else {
		serv.RunService(h)
	}
}
