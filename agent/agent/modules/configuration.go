package modules

import (
	"fmt"
	"net"

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
	TLS Port `json:"tls_port,omitempty"`
}

type CollectorConfiguration struct {
	Integrations map[string]Integration `json:"integrations"`
}

func ReadCollectorConfig() (CollectorConfiguration, error) {
	cnf := CollectorConfiguration{}
	err := utils.ReadJson(configuration.GetCollectorConfigPath(), &cnf)
	if err != nil {
		return cnf, err
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
		newIntegration.TLS.IsListen = false
		newIntegration.TLS.Port = ports.TLS
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

	if proto != "tcp" && proto != "udp" && proto != "tls" {
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
	case "tls":
		integration.TLS.IsListen = isEnabled
		port = integration.TLS.Port
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

	if proto != "tcp" && proto != "udp" && proto != "tls" {
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
	case "tls":
		old = integration.TLS.Port
		integration.TLS.Port = port
	}

	cnf.Integrations[logTyp] = integration
	return old, WriteCollectorConfig(cnf.Integrations, configuration.GetCollectorConfigPath())
}

func IsPortAvailable(port string, proto string, cnf *CollectorConfiguration, currentIntegration string) bool {
	if proto == "tls" {
		proto = "tcp"
	}

	for integration, config := range cnf.Integrations {
		if integration != currentIntegration {
			if config.TCP.Port == port || config.UDP.Port == port || config.TLS.Port == port {
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
