package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/threatwinds/logger"
	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
)

var (
	beatsParser     = BeatsParser{}
	beatsParserOnce sync.Once
	RegexspBeats    = map[configuration.LogType]string{
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
)

type BeatsParser struct{}

func GetBeatsParser() *BeatsParser {
	beatsParserOnce.Do(func() {
		beatsParser = BeatsParser{}
	})
	return &beatsParser
}

func (p *BeatsParser) IdentifySource(log string) (configuration.LogType, error) {
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

func (p *BeatsParser) ProcessData(logBatch interface{}, h *logger.Logger) (map[string][]string, error) {
	classifiedLogs := make(map[string][]string)
	batch, ok := logBatch.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid log batch type")
	}
	for _, log := range batch {
		if logType, err := p.IdentifySource(log); err != nil {
			return nil, err
		} else {
			if logType != "" {
				validatedLog, _, err := validations.ValidateString(log, false)
				if err != nil {
					h.ErrorF("error validating log: %s: %v", log, err)
					continue
				}
				classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], validatedLog)
			}
		}
	}
	return classifiedLogs, nil
}
