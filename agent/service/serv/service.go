package serv

import (
	"context"
	"github.com/kardianos/service"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	pb "github.com/utmstack/UTMStack/agent/service/agent"
	"github.com/utmstack/UTMStack/agent/service/collectors"
	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/database"
	"github.com/utmstack/UTMStack/agent/service/logservice"
	"github.com/utmstack/UTMStack/agent/service/models"
	"github.com/utmstack/UTMStack/agent/service/modules"
	"github.com/utmstack/UTMStack/agent/service/updates"
	"github.com/utmstack/UTMStack/agent/service/utils"
	"google.golang.org/grpc/metadata"
)

type program struct{}

func (p *program) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(_ service.Service) error {
	// TODO: implement this function
	return nil
}

func (p *program) run() {
	utils.InitLogger(config.ServiceLogFile)
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	db := database.GetDB()
	err = db.Migrate(models.Log{})
	if err != nil {
		utils.Logger.ErrorF("error migrating logs table: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")

	go pb.IncidentResponseStream(cnf, ctx)
	go pb.StartPing(cnf, ctx)

	logProcessor := logservice.GetLogProcessor()
	go logProcessor.ProcessLogs(cnf, ctx)

	go modules.StartModules()
	collectors.LogsReader()

	go updates.UpdateDependencies(cnf)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
