package serv

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/utmstack/UTMStack/agent/redline/configuration"
	"github.com/utmstack/UTMStack/agent/redline/protector"
	"github.com/utmstack/UTMStack/agent/redline/utils"
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
	path, err := utils.GetMyPath()
	if err != nil {
		log.Fatalf("Failed to get current path: %v\n", err)
	}

	// Configuring log saving
	var h = utils.CreateLogger(filepath.Join(path, "logs", configuration.SERV_LOG))

	for {
		if utils.CheckIfPathExist(filepath.Join(path, "locks", "setup.lock")) {
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	h.Info("UTMStackRedline started correctly")
	for servName, lockName := range configuration.GetServicesLock() {
		go protector.ProtectService(servName, lockName, h)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

}
