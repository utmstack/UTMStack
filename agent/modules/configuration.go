package modules

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
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

func ReadCollectorConfig() (CollectorConfiguration, error) {
	cnf := CollectorConfiguration{}
	if !utils.CheckIfPathExist(config.CollectorFileName) {
		err := utils.ReadJson(config.CollectorFileNameOld, &cnf)
		if err != nil {
			return CollectorConfiguration{}, err
		}
		err = WriteCollectorConfig(cnf.Integrations, config.CollectorFileName)
		if err != nil {
			return CollectorConfiguration{}, err
		}
		os.Remove(config.CollectorFileNameOld)
	} else {
		err := utils.ReadJson(config.CollectorFileName, &cnf)
		if err != nil {
			return cnf, err
		}
	}

	return cnf, nil
}

func ConfigureCollectorFirstTime() error {
	integrations := make(map[string]Integration)
	for logTyp, ports := range config.ProtoPorts {
		newIntegration := Integration{}
		newIntegration.TCP.IsListen = false
		newIntegration.TCP.Port = ports.TCP
		newIntegration.UDP.IsListen = false
		newIntegration.UDP.Port = ports.UDP
		integrations[string(logTyp)] = newIntegration
	}
	return WriteCollectorConfig(integrations, config.CollectorFileName)
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
	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
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
	return port, WriteCollectorConfig(cnf.Integrations, config.CollectorFileName)
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
	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}
	if changeValid := ValidateChangeInPort(port, logTyp); !changeValid {
		return "", fmt.Errorf("change in port %s protocol %s not allowed for %s or out range %s-%s", port, proto, logTyp, config.PortRangeMin, config.PortRangeMax)
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
	return old, WriteCollectorConfig(cnf.Integrations, config.CollectorFileName)
}

func IsPortAvailable(port string, proto string, cnf *CollectorConfiguration, currentIntegration string) bool {
	for integration, integrationConfig := range cnf.Integrations {
		if integration != currentIntegration {
			if integrationConfig.TCP.Port == port || integrationConfig.UDP.Port == port {
				return false
			}
		}
	}

	listener, err := net.Listen(proto, ":"+port)
	if err != nil {
		return false
	}

	listener.Close()

	return true
}

func WriteCollectorConfig(integrations map[string]Integration, filename string) error {
	fileContent := "{\n    \"integrations\": {\n"
	for name, integration := range integrations {
		fileContent += fmt.Sprintf("        \"%s\": {\n", name)
		if integration.TCP.Port != "" {
			fileContent += fmt.Sprintf("            \"tcp_port\": {\"enabled\": %t, \"value\": \"%s\"},\n", integration.TCP.IsListen, integration.TCP.Port)
		}
		if integration.UDP.Port != "" {
			fileContent += fmt.Sprintf("            \"udp_port\": {\"enabled\": %t, \"value\": \"%s\"},\n", integration.UDP.IsListen, integration.UDP.Port)
		}
		if strings.HasSuffix(fileContent, ",\n") {
			fileContent = fileContent[:len(fileContent)-2] + "\n"
		}
		fileContent += "        },\n"
	}
	if strings.HasSuffix(fileContent, ",\n") {
		fileContent = fileContent[:len(fileContent)-2] + "\n"
	}
	fileContent += "    }\n}\n"

	err := os.WriteFile(filename, []byte(fileContent), 0644)
	if err != nil {
		return err
	}

	return nil
}

func WriteCollectorConfigFromModules(mod []Module, filename string) error {
	integrations := make(map[string]Integration)
	for _, m := range mod {
		integrations[string(m.GetDataType())] = Integration{
			TCP: Port{
				IsListen: m.IsPortListen("tcp"),
				Port:     m.GetPort("tcp"),
			},
			UDP: Port{
				IsListen: m.IsPortListen("udp"),
				Port:     m.GetPort("udp"),
			},
		}
	}
	return WriteCollectorConfig(integrations, filename)
}
