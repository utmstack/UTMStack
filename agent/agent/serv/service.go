package serv

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
	pb "github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/beats"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/conn"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
	"github.com/utmstack/UTMStack/agent/agent/redline"
	"github.com/utmstack/UTMStack/agent/agent/stream"
	"github.com/utmstack/UTMStack/agent/agent/syslog"
	"google.golang.org/grpc/metadata"
)

var h = holmes.New("debug", "UTMStackAgent")

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	// Read config
	cnf, err := configuration.GetCurrentConfig()
	if err != nil {
		fmt.Printf("error getting config: %v", err)
		h.FatalError("error getting config: %v", err)
	}

	// Connect to Agent Manager
	connAgentmanager, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AGENTMANAGERPORT)
	if err != nil {
		fmt.Printf("error connecting to Agent Manager: %v", err)
		h.FatalError("error connecting to Agent Manager: %v", err)
	}
	defer connAgentmanager.Close()
	h.Info("Connection to Agent Manager successful!!!")
	fmt.Printf("Connection to Agent Manager successful!!!")

	// Connect to log-auth-proxy
	connLogServ, err := conn.ConnectToServer(cnf, h, cnf.Server, configuration.AUTHLOGSPORT)
	if err != nil {
		fmt.Printf("error connecting to Log Auth Proxy: %v", err)
		h.FatalError("error connecting to Log Auth Proxy: %v", err)
	}
	defer connLogServ.Close()
	h.Info("Connection to Log Auth Proxy successful!!!")
	fmt.Printf("Connection to Log Auth Proxy successful!!!")

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

	go syslog.SyslogServersUp(h)
	go stream.StartPing(agentClient, ctxAgent, cnf, h)
	go stream.IncidentResponseStream(agentClient, ctxAgent, cnf, h)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

}
