package filters

import (
	"regexp"

	"github.com/utmstack/UTMStack/agent/agent/config"
)

var RegexspBeats = map[config.LogType]string{
	config.LogTypeApacheModule:        `"type":"apache"|"module":"apache"`,
	config.LogTypeLinuxAuditdModule:   `"type":"auditd"|"module":"auditd"`,
	config.LogTypeElasticsearchModule: `"type":"elasticsearch"|"module":"elasticsearch"`,
	config.LogTypeKafkaModule:         `"type":"kafka"|"module":"kafka"`,
	config.LogTypeKibanaModule:        `"type":"kibana"|"module":"kibana"`,
	config.LogTypeLogstashModule:      `"type":"logstash"|"module":"logstash"`,
	config.LogTypeMongodbModule:       `"type":"mongodb"|"module":"mongodb"`,
	config.LogTypeMysqlModule:         `"type":"mysql"|"module":"mysql"`,
	config.LogTypeNginxModule:         `"type":"nginx"|"module":"nginx"`,
	config.LogTypeOsqueryModule:       `"type":"osquery"|"module":"osquery"`,
	config.LogTypePostgresqlModule:    `"type":"postgresql"|"module":"postgresql"`,
	config.LogTypeRedisModule:         `"type":"redis"|"module":"redis"`,
	config.LogTypeLinuxAgent:          `"type":"system"|"module":"system"`,
	config.LogTypeIisModule:           `"type":"iis"|"module":"iis"`,
	config.LogTypeTraefikModule:       `"type":"traefik"|"module":"traefik"`,
	config.LogTypeNatsModule:          `"type":"nats"|"module":"nats"`,
	config.LogTypeHaproxyModule:       `"type":"haproxy"|"module":"haproxy"`,
}

func indentifyBeatLogSource(log string) (config.LogType, error) {
	for logType, regp := range RegexspBeats {
		regExpCompiled, err := regexp.Compile(string(regp))
		if err != nil {
			return "", err
		}
		if regExpCompiled.MatchString(log) {
			return logType, nil
		}
	}
	return config.LogTypeGeneric, nil
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
