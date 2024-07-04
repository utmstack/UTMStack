package serv

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/utmstack/UTMStack/collector-installer/utils"

	"github.com/kardianos/service"
)

type program struct {
	cmdRun  string
	cmdArgs []string
	path    string
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	go func() {
		err := utils.Execute(p.cmdRun, p.path, p.cmdArgs...)
		if err != nil {
			utils.Logger.WriteFatal(fmt.Sprintf("failed to execute command: %s %s: ", p.cmdRun, strings.Join(p.cmdArgs, " ")), err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
