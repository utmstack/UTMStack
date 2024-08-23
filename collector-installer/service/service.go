package serv

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/utmstack/UTMStack/collector-installer/utils"

	"github.com/kardianos/service"
)

var (
	UpdateChann = make(chan bool)
)

type program struct {
	cmdRun   string
	cmdArgs  []string
	path     string
	updating bool
}

func (p *program) Start(s service.Service) error {
	go p.run()
	return nil
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (p *program) run() {
	var ctx context.Context
	var cancel context.CancelFunc

	go func() {
		for {
			if p.updating {
				time.Sleep(15 * time.Second)
				continue
			}

			ctx, cancel = context.WithCancel(context.Background())
			err := utils.ExecuteWithContext(ctx, p.cmdRun, p.path, p.cmdArgs...)
			if err != nil {
				utils.Logger.WriteFatal(fmt.Sprintf("failed to execute command: %s %s: ", p.cmdRun, strings.Join(p.cmdArgs, " ")), err)
			}

			time.Sleep(15 * time.Second)
		}
	}()

	go func() {
		for update := range UpdateChann {
			if update {
				p.updating = true
				cancel()
			} else {
				p.updating = false
			}
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
