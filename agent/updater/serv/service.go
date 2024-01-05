package serv

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kardianos/service"
	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/updater/configuration"
	"github.com/utmstack/UTMStack/agent/updater/updates"
	"github.com/utmstack/UTMStack/agent/updater/utils"
)

var h = holmes.New("debug", configuration.SERV_NAME)

type program struct{}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	for {
		isActive, err := utils.CheckIfServiceIsActive(configuration.GetServAttr()["agent"].ServName)
		if err != nil {
			time.Sleep(time.Second * 5)
			h.Error("error checking if %s service is active: %v", configuration.GetServAttr()["agent"].ServName, err)
			continue
		} else if !isActive {
			time.Sleep(time.Second * 5)
			continue
		}
		break
	}

	path, err := utils.GetMyPath()
	if err != nil {
		h.FatalError("failed to get current path: %v", err)
	}

	var cnf configuration.Config
	err = utils.ReadYAML(filepath.Join(path, "config.yml"), &cnf)
	if err != nil {
		h.FatalError("error reading config.yml: %v", err)
	}

	go updates.UpdateServices(cnf, h)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

}
