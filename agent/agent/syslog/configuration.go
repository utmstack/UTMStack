package syslog

import (
	"fmt"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type ProtocolListen struct {
	Enabled bool   `json:"enabled"`
	UDP     string `json:"UDP,omitempty"`
	TCP     string `json:"TCP,omitempty"`
	TLS     string `json:"TLS,omitempty"`
}

type CollectorConfiguration struct {
	LogCollectorIsenabled bool                      `json:"log_collector_enabled"`
	Integrations          map[string]ProtocolListen `json:"integrations"`
}

func ReadCollectorConfig() (CollectorConfiguration, error) {
	cnf := CollectorConfiguration{}
	path, err := utils.GetMyPath()
	if err != nil {
		return cnf, err
	}

	confPath := filepath.Join(path, "log-collector-config.json")
	err = utils.ReadJson(confPath, &cnf)
	if err != nil {
		return cnf, err
	}
	return cnf, nil
}

func ConfigureCollectorFirstTime() error {
	integrations := make(map[string]ProtocolListen)

	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}
	confPath := filepath.Join(path, "log-collector-config.json")
	for logTyp := range configuration.ProtoPorts {
		if logTyp != configuration.LogTypeSyslog {
			protos := ProtocolListen{}
			protos.Enabled = false
			integrations[string(logTyp)] = protos
		}
	}

	newCnf := CollectorConfiguration{
		LogCollectorIsenabled: false,
		Integrations:          integrations,
	}

	if err := utils.WriteJSON(confPath, &newCnf); err != nil {
		return err
	}
	return nil
}

func CreateNewListener(name string, protos ProtocolListen) error {
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector configuration: ")
	}

	cnf.Integrations[name] = protos

	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}

	confPath := filepath.Join(path, "log-collector-config.json")
	if err := utils.WriteJSON(confPath, &cnf); err != nil {
		return err
	}

	return nil
}

func ChangeListenerConfig(typ string, op string) error {
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector configuration: ")
	}

	switch op {
	case "enable":
		if !cnf.Integrations[string(typ)].Enabled {
			protos := cnf.Integrations[string(typ)]
			protos.Enabled = true
			cnf.Integrations[string(typ)] = protos
		}
	case "disable":
		if cnf.Integrations[string(typ)].Enabled {
			protos := cnf.Integrations[string(typ)]
			protos.Enabled = false
			cnf.Integrations[string(typ)] = protos
		}

	}

	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}
	confPath := filepath.Join(path, "log-collector-config.json")
	if err := utils.WriteJSON(confPath, &cnf); err != nil {
		return err
	}
	return nil
}

func EnableOrDisableLogCollector(op string) error {
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector configuration: ")
	}

	switch op {
	case "enable":
		if !cnf.LogCollectorIsenabled {
			cnf.LogCollectorIsenabled = true
		}
	case "disable":
		if cnf.LogCollectorIsenabled {
			cnf.LogCollectorIsenabled = false
		}
	}

	path, err := utils.GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}
	confPath := filepath.Join(path, "log-collector-config.json")
	if err := utils.WriteJSON(confPath, &cnf); err != nil {
		return err
	}
	return nil
}
