package modules

import (
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
)

const (
	delayCheckSyslogCnfig = 10 * time.Second
)

var (
	moCache = []Module{}
)

type Module interface {
	GetDataType() string
	IsPortListen(proto string) bool
	SetNewPort(proto string, port string)
	GetPort(proto string) string
	EnablePort(proto string)
	DisablePort(proto string)
}

func GetModule(typ configuration.LogType, h *logger.Logger) Module {
	switch configuration.ValidateModuleType(typ) {
	case "syslog":
		return GetSyslogModule(string(typ), configuration.ProtoPorts[typ], h)
	case "netflow":
		return GetNetflowModule(h)
	default:
		return nil
	}
}

func ModulesUp(h *logger.Logger) {
	for {
		time.Sleep(delayCheckSyslogCnfig)
		logCollectorConfig, err := ReadCollectorConfig()
		if err != nil {
			h.Fatal("error reading collector configuration: %v", err)
		}

		for intType, config := range logCollectorConfig.Integrations {
			index := -1
			for i, mod := range moCache {
				if mod.GetDataType() == intType {
					index = i
					break
				}
			}

			if index == -1 {
				newModule := GetModule(configuration.LogType(intType), h)
				moCache = append(moCache, newModule)
				index = len(moCache) - 1
			}

			configs, err := processConfigs(moCache[index], config)
			if err != nil {
				h.ErrorF("error processing configs: %v", err)
				continue
			}

			for proto, conf := range configs {
				changeAllowed := true
				port := ""

				switch proto {
				case "tcp":
					port = config.TCP.Port
				case "udp":
					port = config.UDP.Port
				}

				if port != "" && moCache[index].GetPort(proto) != port {
					changeAllowed = ValidateChangeInPort(port, configuration.LogType(intType))
				}
				if conf[0] {
					moCache[index].DisablePort(proto)
				}
				if changeAllowed {
					moCache[index].SetNewPort(proto, port)
					if conf[1] {
						moCache[index].EnablePort(proto)
					}
				} else {
					h.Info("change in port %s protocol %s not allowed for %s or out range %s-%s", port, proto, intType, configuration.PortRangeMin, configuration.PortRangeMax)
					err := WriteCollectorConfigFromModules(moCache, configuration.GetCollectorConfigPath())
					if err != nil {
						h.ErrorF("error fixing collector configuration: %v", err)
						continue
					}
				}
			}
		}
	}
}

func processConfigs(mod Module, cnf Integration) (map[string][]bool, error) {
	configs := make(map[string][]bool) // first bool is if is necessary kill the port, second bool is if is necessary start the port

	protos := []string{"tcp", "udp"}
	for _, proto := range protos {
		var isEnabled bool
		var port string

		switch proto {
		case "tcp":
			isEnabled = cnf.TCP.IsListen
			port = cnf.TCP.Port
		case "udp":
			isEnabled = cnf.UDP.IsListen
			port = cnf.UDP.Port
		}

		if mod.IsPortListen(proto) && !isEnabled {
			configs[proto] = []bool{true, false}
		} else if !mod.IsPortListen(proto) && isEnabled {
			configs[proto] = []bool{false, true}
		} else if mod.IsPortListen(proto) && isEnabled && mod.GetPort(proto) != port {
			configs[proto] = []bool{true, true}
		} else {
			configs[proto] = []bool{false, false}
		}
	}

	return configs, nil
}

// Return true if the port change is allowed
func ValidateChangeInPort(newPort string, dataType configuration.LogType) bool {
	for _, logType := range configuration.ProhibitedPortsChange {
		if logType == dataType {
			return false
		}
	}
	return configuration.PortRangeMin <= newPort && newPort <= configuration.PortRangeMax
}
