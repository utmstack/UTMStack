package serv

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/redline/constants"
	"github.com/utmstack/UTMStack/agent/redline/protector"
	"github.com/utmstack/UTMStack/agent/redline/utils"
)

var h = holmes.New("debug", constants.SERV_NAME)

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
		h.FatalError("Failed to get current path: %v", err)
	}

	for {
		if utils.CheckIfPathExist(filepath.Join(path, "locks", "setup.lock")) {
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	h.Info("UTMStackRedline started correctly")
	for servName, lockName := range constants.GetServicesLock() {
		go protector.ProtectService(servName, lockName, h)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

}
