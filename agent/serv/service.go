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
	"github.com/utmstack/UTMStack/agent/database"
	"github.com/utmstack/UTMStack/agent/logservice"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/modules"
	"github.com/utmstack/UTMStack/agent/updates"
	"github.com/utmstack/UTMStack/agent/utils"
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

	if utils.CheckIfPathExist(config.IntegrationCertPath) && utils.CheckIfPathExist(config.IntegrationKeyPath) {
		err = utils.ValidateIntegrationCertificates(config.IntegrationCertPath, config.IntegrationKeyPath)
		if err != nil {
			utils.Logger.ErrorF("TLS certificates are invalid: %v", err)
			utils.Logger.Info("TLS functionality will be disabled. Use 'load-tls-certs' command to fix this.")
		} else {
			utils.Logger.Info("TLS certificates are valid and ready for use")
		}
	} else {
		utils.Logger.Info("No TLS certificates found. TLS functionality is disabled.")
		utils.Logger.Info("Use 'load-tls-certs' command to load your certificates.")
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

	go pb.UpdateAgent(cnf, ctx)
	go modules.StartModules()
	collectors.LogsReader()

	go updates.UpdateDependencies(cnf)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
