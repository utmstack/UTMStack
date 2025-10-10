package modules

import (
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
)

const (
	delayCheckSyslogCnfig = 10 * time.Second
)

var (
	moCache = make([]Module, 0, 10)
)

type Module interface {
	GetDataType() string
	IsPortListen(proto string) bool
	SetNewPort(proto string, port string)
	GetPort(proto string) string
	EnablePort(proto string, enableTLS bool) error
	DisablePort(proto string)
}

func GetModule(typ string) Module {
	switch config.ValidateModuleType(typ) {
	case "syslog":
		return GetSyslogModule(typ, config.ProtoPorts[config.DataType(typ)])
	case "netflow":
		return GetNetflowModule()
	default:
		return nil
	}
}

func StartModules() {
	for {
		time.Sleep(delayCheckSyslogCnfig)
		logCollectorConfig, err := ReadCollectorConfig()
		if err != nil {
			utils.Logger.Fatal("error reading collector config: %v", err)
		}

		for intType, cnf := range logCollectorConfig.Integrations {
			index := -1
			for i, mod := range moCache {
				if mod.GetDataType() == intType {
					index = i
					break
				}
			}

			if index == -1 {
				newModule := GetModule(intType)
				if newModule == nil {
					utils.Logger.LogF(100, "error getting module %s", intType)
					continue
				}
				moCache = append(moCache, newModule)
				index = len(moCache) - 1
			}

			configs, err := processConfigs(moCache[index], cnf)
			if err != nil {
				utils.Logger.ErrorF("error processing configs: %v", err)
				continue
			}

			for proto, conf := range configs {
				changeAllowed := true
				port := ""

				switch proto {
				case "tcp":
					port = cnf.TCP.Port
				case "udp":
					port = cnf.UDP.Port
				}

				if port != "" && moCache[index].GetPort(proto) != port {
					changeAllowed = ValidateChangeInPort(port, intType)
				}
				if conf[0] {
					moCache[index].DisablePort(proto)
				}
				if changeAllowed {
					moCache[index].SetNewPort(proto, port)
					if conf[1] {
						enableTLS := proto == "tcp" && cnf.TCP.TLSEnabled

						err := moCache[index].EnablePort(proto, enableTLS)
						if err != nil {
							utils.Logger.ErrorF("error enabling port for %s %s: %v", intType, proto, err)
						}
					}
				} else {
					utils.Logger.Info("change in port %s protocol %s not allowed for %s or out range %s-%s", port, proto, intType, config.PortRangeMin, config.PortRangeMax)
					err := WriteCollectorConfigFromModules(moCache, config.CollectorFileName)
					if err != nil {
						utils.Logger.ErrorF("error fixing collector config: %v", err)
						continue
					}
				}
			}
		}
	}
}

func processConfigs(mod Module, cnf Integration) (map[string][]bool, error) {
	configs := make(map[string][]bool) // first bool is if is necessary kill the port, second bool is if is necessary start the port

	protocols := []string{"tcp", "udp"}
	for _, protocol := range protocols {
		var isEnabled bool
		var port string

		switch protocol {
		case "tcp":
			isEnabled = cnf.TCP.IsListen
			port = cnf.TCP.Port
		case "udp":
			isEnabled = cnf.UDP.IsListen
			port = cnf.UDP.Port
		}

		if mod.IsPortListen(protocol) && !isEnabled {
			configs[protocol] = []bool{true, false}
		} else if !mod.IsPortListen(protocol) && isEnabled {
			configs[protocol] = []bool{false, true}
		} else if mod.IsPortListen(protocol) && isEnabled && mod.GetPort(protocol) != port {
			configs[protocol] = []bool{true, true}
		} else {
			configs[protocol] = []bool{false, false}
		}
	}

	return configs, nil
}

// ValidateChangeInPort returns true if the port change is allowed
func ValidateChangeInPort(newPort string, dataType string) bool {
	for _, logType := range config.ProhibitedPortsChange {
		if string(logType) == dataType {
			return false
		}
	}
	return config.PortRangeMin <= newPort && newPort <= config.PortRangeMax
}
