package modules

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

type Port struct {
	IsListen   bool   `json:"enabled"`
	Port       string `json:"value"`
	TLSEnabled bool   `json:"tls_enabled,omitempty"`
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
	err := utils.ReadJson(config.CollectorFileName, &cnf)
	if err != nil {
		return cnf, err
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

func ChangeIntegrationStatus(logTyp string, proto string, isEnabled bool, tlsOptions ...bool) (string, error) {
	var port string
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		integration.TCP.IsListen = isEnabled
		port = integration.TCP.Port

		// Handle TLS configuration if specified
		if len(tlsOptions) > 0 && isEnabled {
			if tlsOptions[0] {
				if !utils.CheckIfPathExist(config.IntegrationCertPath) || !utils.CheckIfPathExist(config.IntegrationKeyPath) {
					return "", fmt.Errorf("TLS certificates not found. Please load certificates first")
				}
				// Enable TLS
				integration.TCP.TLSEnabled = true
				mod := GetModule(logTyp)
				if mod != nil && mod.IsPortListen(proto) {
					mod.DisablePort(proto)
					time.Sleep(100 * time.Millisecond)
					err := mod.EnablePort(proto, true)
					if err != nil {
						return "", fmt.Errorf("error enabling TLS on running module: %v", err)
					}
				}
			} else {
				// Disable TLS
				integration.TCP.TLSEnabled = false
				mod := GetModule(logTyp)
				if mod != nil && mod.IsPortListen(proto) {
					mod.DisablePort(proto)
					time.Sleep(100 * time.Millisecond)
					err := mod.EnablePort(proto, false)
					if err != nil {
						return "", fmt.Errorf("error disabling TLS on running module: %v", err)
					}
				}
			}
		}

		// Auto-disable TLS when disabling integration
		if !isEnabled {
			integration.TCP.TLSEnabled = false
		}

	case "udp":
		integration.UDP.IsListen = isEnabled
		port = integration.UDP.Port

		// TLS validation for UDP
		if len(tlsOptions) > 0 && tlsOptions[0] {
			return "", fmt.Errorf("TLS is not supported for UDP protocol. Use TCP for TLS connections")
		}

	default:
		return "", fmt.Errorf("invalid protocol: %s", proto)
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
	default:
		return "", fmt.Errorf("invalid protocol: %s", proto)
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
			fileContent += fmt.Sprintf("            \"tcp_port\": {\"enabled\": %t, \"value\": \"%s\"", integration.TCP.IsListen, integration.TCP.Port)
			if integration.TCP.TLSEnabled {
				fileContent += fmt.Sprintf(", \"tls_enabled\": %t", integration.TCP.TLSEnabled)
			}
			fileContent += "},\n"
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
		integrations[m.GetDataType()] = Integration{
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

func EnableTLSForIntegration(logTyp string, proto string) (string, error) {
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}

	integration := cnf.Integrations[logTyp]
	var port string

	switch proto {
	case "tcp":
		if integration.TCP.Port == "" {
			return "", fmt.Errorf("TCP port not configured for %s", logTyp)
		}
		port = integration.TCP.Port
		integration.TCP.TLSEnabled = true

		mod := GetModule(logTyp)
		if mod != nil && mod.IsPortListen(proto) {
			mod.DisablePort(proto)
			time.Sleep(100 * time.Millisecond)
			err := mod.EnablePort(proto, true)
			if err != nil {
				return port, fmt.Errorf("error enabling TLS on running module: %v", err)
			}
		}
	case "udp":
		return "", fmt.Errorf("TLS not supported for UDP protocol")
	default:
		return "", fmt.Errorf("invalid protocol: %s", proto)
	}

	cnf.Integrations[logTyp] = integration
	return port, WriteCollectorConfig(cnf.Integrations, config.CollectorFileName)
}

func DisableTLSForIntegration(logTyp string, proto string) error {
	cnf, err := ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector config: %v", err)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		integration.TCP.TLSEnabled = false

		mod := GetModule(logTyp)
		if mod != nil && mod.IsPortListen(proto) {
			mod.DisablePort(proto)
			time.Sleep(100 * time.Millisecond)
			err := mod.EnablePort(proto, false)
			if err != nil {
				return fmt.Errorf("error disabling TLS on running module: %v", err)
			}
		}
	case "udp":
		return fmt.Errorf("TLS not supported for UDP protocol")
	default:
		return fmt.Errorf("invalid protocol: %s", proto)
	}

	cnf.Integrations[logTyp] = integration
	return WriteCollectorConfig(cnf.Integrations, config.CollectorFileName)
}
