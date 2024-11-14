package serv

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/kardianos/service"
	pb "github.com/utmstack/UTMStack/agent/service/agent"
	"github.com/utmstack/UTMStack/agent/service/beats"
	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/conn"
	"github.com/utmstack/UTMStack/agent/service/database"
	"github.com/utmstack/UTMStack/agent/service/logservice"
	"github.com/utmstack/UTMStack/agent/service/models"
	"github.com/utmstack/UTMStack/agent/service/modules"
	"github.com/utmstack/UTMStack/agent/service/updates"
	"github.com/utmstack/UTMStack/agent/service/utils"
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
	path := utils.GetMyPath()
	utils.InitLogger(filepath.Join(path, "logs", config.SERV_LOG))
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	db := database.GetDB()
	err = db.Migrate(models.Log{})
	if err != nil {
		utils.Logger.ErrorF("error migrating logs table: %v", err)
	}

	err = conn.EstablishConnectionsFistTime(cnf)
	if err != nil {
		utils.Logger.ErrorF("error establishing connections: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.AgentKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.AgentID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "agent")

	go pb.IncidentResponseStream(cnf, ctx)
	go pb.StartPing(cnf, ctx)

	logp := logservice.GetLogProcessor()
	go logp.ProcessLogs(cnf, ctx)

	go modules.ModulesUp()
	beats.BeatsLogsReader()

	go updates.UpdateDependencies(cnf)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
