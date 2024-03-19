package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/modules"
	"github.com/utmstack/UTMStack/agent/agent/serv"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc/metadata"
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
			h.Info("Installing UTMStack Agent service...")

			// Generate Certificates
			certsPath := filepath.Join(path, "certs")
			err = utils.CreatePathIfNotExist(certsPath)
			if err != nil {
				h.Fatal("error creating path: %s", err)
			}

			err = utils.GenerateCerts(certsPath)
			if err != nil {
				h.Fatal("error generating certificates: %v", err)
			}

			cnf, utmKey := configuration.GetInitialConfig()

			// Connect to Agent Manager
			conn, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
			if err != nil {
				h.Fatal("error connecting to Agent Manager: %v", err)
			}
			defer conn.Close()
			h.Info("Connection to Agent Manager successful!!!")

			// Register Agent
			if err = pb.RegisterAgent(conn, cnf, utmKey, h); err != nil {
				h.Fatal("%v", err)
			}

			// Write config in config.yml
			if err = configuration.SaveConfig(cnf); err != nil {
				h.Fatal("error writing config file: %v", err)
			}

			if err = modules.ConfigureCollectorFirstTime(); err != nil {
				h.Fatal("error configuring syslog server: %v", err)
			}

			// Install Beats
			if err = beats.InstallBeats(*cnf, h); err != nil {
				h.Fatal("error installing beats: %v", err)
			}

			serv.InstallService(h)
			h.Info("UTMStack Agent service installed correctly")

		case "send-log":
			msg := os.Args[2]
			logp := logservice.GetLogProcessor()

			// Read config
			cnf, err := configuration.GetCurrentConfig()
			if err != nil {
				os.Exit(0)
			}

			// Connect to log-auth-proxy
			connLogServ, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AUTHLOGSPORT)
			if err != nil {
				h.ErrorF("error connecting to Log Auth Proxy: %v", err)
			}
			defer connLogServ.Close()

			// Create a client for LogService
			logClient := logservice.NewLogServiceClient(connLogServ)
			ctxLog, cancelLog := context.WithCancel(context.Background())
			defer cancelLog()
			ctxLog = metadata.AppendToOutgoingContext(ctxLog, "agent-key", cnf.AgentKey)

			err = logp.ProcessLogsWithHighPriority(msg, logClient, ctxLog, cnf, h)
			if err != nil {
				h.ErrorF("error sending High Priority Log to Log Auth Proxy: %v", err)
			}
		case "enable-integration", "disable-integration":
			integration := os.Args[2]
			proto := os.Args[3]

			port, err := modules.ChangeIntegrationStatus(integration, proto, (arg == "enable-integration"))
			if err != nil {
				fmt.Printf("error trying to %s: %v", arg, err)
				h.ErrorF("error trying to %s: %v", arg, err)
				os.Exit(0)
			}
			fmt.Printf("%s %s done correctly in port %s %s", arg, integration, port, proto)
			time.Sleep(5 * time.Second)

		case "change-port":
			integration := os.Args[2]
			proto := os.Args[3]
			port := os.Args[4]

			old, err := modules.ChangePort(integration, proto, port)
			if err != nil {
				fmt.Printf("error trying to change port: %v", err)
				h.ErrorF("error trying to change port: %v", err)
				os.Exit(0)
			}
			fmt.Printf("change port done correctly from %s to %s in %s for %s integration", old, port, proto, integration)
			time.Sleep(5 * time.Second)

		case "uninstall":
			h.Info("Uninstalling UTMStack Agent service...")

			// Read config
			cnf, err := configuration.GetCurrentConfig()
			if err != nil {
				h.Fatal("error getting config: %v", err)
			}

			// Connect to Agent Manager
			conn, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
			if err != nil {
				h.ErrorF("error connecting to Agent Manager: %v", err)
			} else {
				h.Info("Connection to Agent Manager successful!!!")

				// Delete agent
				if err = pb.DeleteAgent(conn, cnf, h); err != nil {
					h.ErrorF("error deleting agent: %v", err)
				}
			}
			defer conn.Close()

			// Uninstall Beats
			if err = beats.UninstallBeats(h); err != nil {
				h.Fatal("error uninstalling beats: %v", err)
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
