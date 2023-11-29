package filters

import (
	"regexp"

	"github.com/utmstack/UTMStack/agent/agent/configuration"
)

var RegexspBeats = map[configuration.LogType]string{
	configuration.LogTypeApacheModule:        `"type":"apache"|"module":"apache"`,
	configuration.LogTypeLinuxAuditdModule:   `"type":"auditd"|"module":"auditd"`,
	configuration.LogTypeElasticsearchModule: `"type":"elasticsearch"|"module":"elasticsearch"`,
	configuration.LogTypeKafkaModule:         `"type":"kafka"|"module":"kafka"`,
	configuration.LogTypeKibanaModule:        `"type":"kibana"|"module":"kibana"`,
	configuration.LogTypeLogstashModule:      `"type":"logstash"|"module":"logstash"`,
	configuration.LogTypeMongodbModule:       `"type":"mongodb"|"module":"mongodb"`,
	configuration.LogTypeMysqlModule:         `"type":"mysql"|"module":"mysql"`,
	configuration.LogTypeNginxModule:         `"type":"nginx"|"module":"nginx"`,
	configuration.LogTypeOsqueryModule:       `"type":"osquery"|"module":"osquery"`,
	configuration.LogTypePostgresqlModule:    `"type":"postgresql"|"module":"postgresql"`,
	configuration.LogTypeRedisModule:         `"type":"redis"|"module":"redis"`,
	configuration.LogTypeLinuxAgent:          `"type":"system"|"module":"system"`,
	configuration.LogTypeIisModule:           `"type":"iis"|"module":"iis"`,
	configuration.LogTypeTraefikModule:       `"type":"traefik"|"module":"traefik"`,
	configuration.LogTypeNatsModule:          `"type":"nats"|"module":"nats"`,
	configuration.LogTypeHaproxyModule:       `"type":"haproxy"|"module":"haproxy"`,
}

func indentifyBeatLogSource(log string) (configuration.LogType, error) {
	for logType, regp := range RegexspBeats {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType, nil
		}
	}
	return configuration.LogTypeGeneric, nil
}

func ProcessBeatData(logBatch []string) (map[string][]string, error) {
	classifiedLogs := make(map[string][]string)
	for _, log := range logBatch {
		if logType, err := indentifyBeatLogSource(log); err != nil {
			return nil, err
		} else {
			if logType != "" {
				classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], log)
			}
		}
	}
	return classifiedLogs, nil
}
