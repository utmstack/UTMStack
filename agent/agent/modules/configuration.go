package modules

import (
	"fmt"
	"net"
	"os"

	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type Port struct {
	IsListen bool   `json:"enabled"`
	Port     string `json:"value"`
}

type Integration struct {
	TCP Port `json:"tcp_port,omitempty"`
	UDP Port `json:"udp_port,omitempty"`
}

type CollectorConfiguration struct {
	Integrations map[string]Integration `json:"integrations"`
}

type ProtocolListenOld struct {
	Enabled bool   `json:"enabled"`
	UDP     string `json:"UDP,omitempty"`
	TCP     string `json:"TCP,omitempty"`
	TLS     string `json:"TLS,omitempty"`
}

type CollectorConfigurationOld struct {
	LogCollectorIsenabled bool                         `json:"log_collector_enabled"`
	Integrations          map[string]ProtocolListenOld `json:"integrations"`
}

func ReadCollectorConfig() (CollectorConfiguration, error) {
	cnf := CollectorConfiguration{}
	if !utils.CheckIfPathExist(configuration.GetCollectorConfigPath()) {
		cnfOld := CollectorConfigurationOld{}
		err := utils.ReadJson(configuration.GetCollectorConfigPathOld(), &cnfOld)
		if err != nil {
			return CollectorConfiguration{}, err
		}
		cnf = MigrateOldConfig(cnfOld)
		err = WriteCollectorConfig(cnf.Integrations, configuration.GetCollectorConfigPath())
		if err != nil {
			return CollectorConfiguration{}, err
		}
		os.Remove(configuration.GetCollectorConfigPathOld())
	} else {
		err := utils.ReadJson(configuration.GetCollectorConfigPath(), &cnf)
		if err != nil {
			return cnf, err
		}
	}

	return cnf, nil
}

func ConfigureCollectorFirstTime() error {
	integrations := make(map[string]Integration)
	for logTyp, ports := range configuration.ProtoPorts {
		newIntegration := Integration{}
		newIntegration.TCP.IsListen = false
		newIntegration.TCP.Port = ports.TCP
		newIntegration.UDP.IsListen = false
		newIntegration.UDP.Port = ports.UDP
		integrations[string(logTyp)] = newIntegration
	}
	return WriteCollectorConfig(integrations, configuration.GetCollectorConfigPath())
}

func ChangeIntegrationStatus(logTyp string, proto string, isEnabled bool) (string, error) {
	var port string
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if proto != "tcp" && proto != "udp" {
		return "", fmt.Errorf("invalid protocol: %s", proto)
	}
	if valid := configuration.ValidateModuleType(configuration.LogType(logTyp)); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		integration.TCP.IsListen = isEnabled
		port = integration.TCP.Port
	case "udp":
		integration.UDP.IsListen = isEnabled
		port = integration.UDP.Port
	}

	cnf.Integrations[logTyp] = integration
	return port, WriteCollectorConfig(cnf.Integrations, configuration.GetCollectorConfigPath())
}

func ChangePort(logTyp string, proto string, port string) (string, error) {
	var old string
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if proto != "tcp" && proto != "udp" {
		return "", fmt.Errorf("invalid protocol: %s", proto)
	}
	if valid := configuration.ValidateModuleType(configuration.LogType(logTyp)); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}
	if changeValid := ValidateChangeInPort(port, configuration.LogType(logTyp)); !changeValid {
		return "", fmt.Errorf("change in port %s protocol %s not allowed for %s or out range %s-%s", port, proto, logTyp, configuration.PortRangeMin, configuration.PortRangeMax)
	}
	if !IsPortAvailable(port, proto, &cnf, logTyp) {
		return "", fmt.Errorf("port %s is already in use", port)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		old = integration.TCP.Port
		integration.TCP.Port = port
	case "udp":
		old = integration.UDP.Port
		integration.UDP.Port = port
	}

	cnf.Integrations[logTyp] = integration
	return old, WriteCollectorConfig(cnf.Integrations, configuration.GetCollectorConfigPath())
}

func IsPortAvailable(port string, proto string, cnf *CollectorConfiguration, currentIntegration string) bool {
	for integration, config := range cnf.Integrations {
		if integration != currentIntegration {
			if config.TCP.Port == port || config.UDP.Port == port {
				return false
			}
		}
	}

	ln, err := net.Listen(proto, ":"+port)
	if err != nil {
		return false
	}
	ln.Close()

	return true
}

func MigrateOldConfig(old CollectorConfigurationOld) CollectorConfiguration {
	integrations := make(map[string]Integration)
	for logTyp, ports := range configuration.ProtoPorts {
		newIntegration := Integration{}
		if logTyp == "syslog" && old.LogCollectorIsenabled {
			newIntegration.TCP.IsListen = true
		} else {
			newIntegration.TCP.IsListen = false
		}
		newIntegration.TCP.Port = ports.TCP
		newIntegration.UDP.IsListen = false
		newIntegration.UDP.Port = ports.UDP
		integrations[string(logTyp)] = newIntegration
	}

	for logTyp, port := range old.Integrations {
		if _, ok := integrations[logTyp]; ok {
			if old.LogCollectorIsenabled && port.Enabled {
				integration := integrations[logTyp]
				integration.TCP.IsListen = true
				integration.UDP.IsListen = false
				integrations[logTyp] = integration
			}
		}
	}

	return CollectorConfiguration{Integrations: integrations}
}
