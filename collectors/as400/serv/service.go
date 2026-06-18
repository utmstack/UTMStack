package serv

import (
	"context"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kardianos/service"

	pb "github.com/utmstack/UTMStack/collectors/as400/agent"
	collectors "github.com/utmstack/UTMStack/collectors/as400/collector"
	"github.com/utmstack/UTMStack/collectors/as400/config"
	"github.com/utmstack/UTMStack/collectors/as400/logservice"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
	"google.golang.org/grpc/metadata"
)

type program struct {
	as400 *collectors.AS400Collector
}

func (p *program) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(_ service.Service) error {
	if p.as400 != nil {
		utils.Logger.Info("Stopping AS400 Collector...")
		return p.as400.Stop()
	}
	return nil
}

func (p *program) run() {
	utils.InitLogger(config.ServiceLogFile)
	cnf, err := config.GetCurrentConfig()
	if err != nil {
		utils.Logger.Fatal("error getting config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "key", cnf.CollectorKey)
	ctx = metadata.AppendToOutgoingContext(ctx, "id", strconv.Itoa(int(cnf.CollectorID)))
	ctx = metadata.AppendToOutgoingContext(ctx, "type", "collector")

	go pb.StartPing(cnf, ctx)

	logProcessor := logservice.GetLogProcessor()
	go logProcessor.ProcessLogs(cnf, ctx)

	// Start AS400 Collector; it watches the local config file and hot-reloads.
	p.as400 = collectors.NewAS400Collector()
	if err := p.as400.Start(ctx); err != nil {
		utils.Logger.Fatal("error starting AS400 collector: %v", err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
