package serv

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
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

	db := database.GetDB()
	err = db.Migrate(models.Log{})
	if err != nil {
		utils.Logger.ErrorF("error migrating logs table: %v", err)
	}

	ensureUpdaterServiceInstalled()

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

func ensureUpdaterServiceInstalled() {
	isInstalled, err := utils.CheckIfServiceIsInstalled("UTMStackUpdater")
	if err != nil {
		utils.Logger.ErrorF("error checking if updater service is installed: %v", err)
		return
	}

	if isInstalled {
		utils.Logger.Info("updater service is already installed")
		return
	}

	utils.Logger.Info("updater service not found, installing...")

	updaterPath := filepath.Join(utils.GetMyPath(), fmt.Sprintf(config.UpdaterFile, ""))

	if !utils.CheckIfPathExist(updaterPath) {

		cnf, err := config.GetCurrentConfig()
		if err != nil {
			utils.Logger.ErrorF("error getting config to download updater: %v", err)
			return
		}

		updaterBinary := fmt.Sprintf(config.UpdaterFile, "")
		if err := utils.DownloadFile(fmt.Sprintf(config.DependUrl, cnf.Server, config.DependenciesPort, updaterBinary), map[string]string{}, updaterBinary, utils.GetMyPath(), cnf.SkipCertValidation); err != nil {
			utils.Logger.ErrorF("error downloading updater binary: %v", err)
			return
		}

		if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
			if err := utils.Execute("chmod", utils.GetMyPath(), "755", updaterBinary); err != nil {
				utils.Logger.ErrorF("error setting permissions on updater: %v", err)
				return
			}
		}

		utils.Logger.Info("updater binary downloaded successfully")
	}

	err = utils.Execute(updaterPath, utils.GetMyPath(), "install")
	if err != nil {
		utils.Logger.ErrorF("error installing updater service: %v", err)
		return
	}

	utils.Logger.Info("updater service installed successfully")
}
