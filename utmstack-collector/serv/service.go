package serv

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kardianos/service"

	pb "github.com/utmstack/UTMStack/utmstack-collector/agent"
	"github.com/utmstack/UTMStack/utmstack-collector/collector"
	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/database"
	"github.com/utmstack/UTMStack/utmstack-collector/logservice"
	"github.com/utmstack/UTMStack/utmstack-collector/models"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
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
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.CollectorKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.CollectorID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "collector")

	go pb.StartPing(cnf, ctx)

	logProcessor := logservice.GetLogProcessor()
	go logProcessor.ProcessLogs(cnf, ctx)

	// Start UTMStack log collector
	dockerConfig := collector.DefaultConfig()
	dockerCollector, err := collector.NewDockerCollector(dockerConfig)
	if err != nil {
		utils.Logger.ErrorF("failed to create UTMStack Collector: %v", err)
	} else {
		go func() {
			if err := dockerCollector.Start(); err != nil {
				utils.Logger.ErrorF("failed to start UTMStack Collector: %v", err)
			}
		}()
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
