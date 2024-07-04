package parser

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/threatwinds/validations"
	"github.com/utmstack/UTMStack/agent/agent/config"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

var (
	beatsParser     = BeatsParser{}
	beatsParserOnce sync.Once
	RegexspBeats    = map[config.LogType]string{
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
)

type BeatsParser struct{}

func GetBeatsParser() *BeatsParser {
	beatsParserOnce.Do(func() {
		beatsParser = BeatsParser{}
	})
	return &beatsParser
}

func (p *BeatsParser) IdentifySource(log string) (config.LogType, error) {
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

func (p *BeatsParser) ProcessData(logBatch interface{}) (map[string][]string, error) {
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
					utils.Logger.ErrorF("error validating log: %s: %v", log, err)
					continue
				}
				classifiedLogs[string(logType)] = append(classifiedLogs[string(logType)], validatedLog)
			}
		}
	}
	return classifiedLogs, nil
}
