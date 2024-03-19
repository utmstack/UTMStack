package serv

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/kardianos/service"
	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/modules"
	"github.com/utmstack/UTMStack/agent/agent/redline"
	"github.com/utmstack/UTMStack/agent/agent/stream"
	"github.com/utmstack/UTMStack/agent/agent/utils"
	"google.golang.org/grpc/metadata"
)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	// Get current path
	path, err := utils.GetMyPath()
	if err != nil {
		log.Fatalf("Failed to get current path: %v", err)
	}

	// Configuring log saving
	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))
	// Read config
	cnf, err := configuration.GetCurrentConfig()
	if err != nil {
		h.Fatal("error getting config: %v", err)
	}

	// Connect to Agent Manager
	connAgentmanager, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
	if err != nil {
		h.Fatal("error connecting to Agent Manager: %v", err)
	}
	defer connAgentmanager.Close()
	h.Info("Connection to Agent Manager successful!!!")

	// Connect to log-auth-proxy
	connLogServ, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AUTHLOGSPORT)
	if err != nil {
		h.Fatal("error connecting to Log Auth Proxy: %v", err)
	}
	defer connLogServ.Close()
	h.Info("Connection to Log Auth Proxy successful!!!")

	// Create a client for AgentService
	agentClient := pb.NewAgentServiceClient(connAgentmanager)
	ctxAgent, cancelAgent := context.WithCancel(context.Background())
	defer cancelAgent()
	ctxAgent = metadata.AppendToOutgoingContext(ctxAgent, "agent-key", cnf.AgentKey)
	ctxAgent = metadata.AppendToOutgoingContext(ctxAgent, "agent-id", strconv.Itoa(int(cnf.AgentID)))

	// Create a client for LogService
	logClient := logservice.NewLogServiceClient(connLogServ)
	ctxLog, cancelLog := context.WithCancel(context.Background())
	defer cancelLog()
	ctxLog = metadata.AppendToOutgoingContext(ctxLog, "agent-key", cnf.AgentKey)

	logp := logservice.GetLogProcessor()
	go logp.ProcessLogs(logClient, ctxLog, cnf, h)

	beats.BeatsLogsReader(h)
	go redline.CheckRedlineService(h)

	go modules.ModulesUp(h)
	go stream.StartPing(agentClient, ctxAgent, cnf, h)
	go stream.IncidentResponseStream(agentClient, ctxAgent, cnf, h)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
