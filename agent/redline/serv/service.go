package serv

import (
	"os"
	"os/signal"
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
checkall:
	for {
		for servName := range constants.GetServicesLock() {
			isActive, err := utils.CheckIfServiceIsActive(servName)
			if err != nil {
				time.Sleep(time.Second * 5)
				h.Error("error checking if %s service is active: %v", servName, err)
				continue checkall
			} else if !isActive {
				time.Sleep(time.Second * 5)
				continue checkall
			}
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
