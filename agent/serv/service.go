package serv

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kardianos/service"
	pb "github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/collectors"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/modules"
	"github.com/utmstack/UTMStack/agent/updates"
	"github.com/utmstack/UTMStack/agent/utils"
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
	utils.InitLogger(config.ServiceLogFile)
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	CleanOldServices(cnf)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")

	go pb.IncidentResponseStream(cnf, ctx)
	go pb.StartPing(cnf, ctx)

	logProcessor := logservice.GetLogProcessor()
	go logProcessor.ProcessLogs(cnf, ctx)

	go pb.UpdateAgent(cnf, ctx)
	go modules.StartModules()
	collectors.LogsReader()

	go updates.UpdateDependencies(cnf)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
