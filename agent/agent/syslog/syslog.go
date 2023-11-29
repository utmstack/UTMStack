package syslog

import (
	"fmt"
	"time"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/logservice"
)

const (
	delayCheckSyslogCnfig = 10 * time.Second
)

func SyslogServersUp(h *holmes.Logger) {
	var collectorGotUpFirstTime bool
	syslogServers := []SyslogServer{}

	for {
		time.Sleep(delayCheckSyslogCnfig)
		logCollectorConfig, err := ReadCollectorConfig()
		if err != nil {
			fmt.Printf("error reading collector configuration: %v", err)
			h.FatalError("error reading collector configuration: %v", err)
		}
		if logCollectorConfig.LogCollectorIsenabled {
			if !collectorGotUpFirstTime {
				syslogServers = append(syslogServers, NewSyslogServer(
					string(configuration.LogTypeSyslog),
					true,
					configuration.ProtoPorts[configuration.LogTypeSyslog],
					"0.0.0.0",
				))

				collectorGotUpFirstTime = true
			}

			for intType, config := range logCollectorConfig.Integrations {
				var existsServer bool
				var index int
				for i, serv := range syslogServers {
					if serv.DataType == intType {
						existsServer = true
						index = i
					}
				}
				if config.Enabled && !existsServer {
					integrationPorts := configuration.ProtoPort{}
					var isIntegrationDefault bool
					for typ, ports := range configuration.ProtoPorts {
						if string(typ) == intType {
							isIntegrationDefault = true
							integrationPorts = ports
						}
					}

					if !isIntegrationDefault {
						integrationPorts = configuration.ProtoPort{
							TCP: config.TCP,
							UDP: config.UDP,
							TLS: config.TLS,
						}
					}

					syslogServers = append(syslogServers, NewSyslogServer(
						intType,
						true,
						integrationPorts,
						"0.0.0.0",
					))
				} else if !config.Enabled && existsServer {
					syslogServers[index].Kill(h)
					syslogServers = append(syslogServers[:index], syslogServers[index+1:]...)
				}

			}

			for i := range syslogServers {
				if !syslogServers[i].IsListen {
					syslogServers[i].SetHandler(logservice.LogQueue)
					syslogServers[i].Listen(h)
					syslogServers[i].IsListen = true
				}
			}
		} else {
			if collectorGotUpFirstTime {
				for index := range syslogServers {
					syslogServers[index].Kill(h)
				}
				syslogServers = []SyslogServer{}
				collectorGotUpFirstTime = false
			}
		}
	}
}
