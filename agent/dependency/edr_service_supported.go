//go:build windows || linux

package dependency

import (
	"errors"
	"fmt"

	"github.com/kardianos/service"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
)

type edrServiceProbe struct{}

func (edrServiceProbe) Start(service.Service) error { return nil }
func (edrServiceProbe) Stop(service.Service) error  { return nil }

func edrServiceStatus() (edrServiceState, error) {
	s, err := service.New(edrServiceProbe{}, &service.Config{Name: edrconfig.ServiceName})
	if err != nil {
		return edrServiceState{}, err
	}
	st, err := s.Status()
	if errors.Is(err, service.ErrNotInstalled) {
		return edrServiceState{}, nil
	}
	if err != nil {
		if edrFailedServiceInstalled() {
			return edrServiceState{installed: true}, nil
		}
		return edrServiceState{}, err
	}
	switch st {
	case service.StatusRunning:
		return edrServiceState{installed: true, running: true}, nil
	case service.StatusStopped:
		return edrServiceState{installed: true}, nil
	default:
		return edrServiceState{}, fmt.Errorf("EDR service is in an unknown state")
	}
}
