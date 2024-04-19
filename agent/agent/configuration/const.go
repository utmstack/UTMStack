package configuration

import (
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

const REPLACE_KEY string = ""

const (
	AGENTMANAGERPORT      string = "9000"
	AUTHLOGSPORT          string = "50051"
	BatchCapacity                = 100
	MaxConnectionTime            = 120 * time.Second
	InitialReconnectDelay        = 10 * time.Second
	MaxReconnectDelay            = 120 * time.Second

	SERV_NAME            = "UTMStackAgent"
	SERV_LOG             = "utmstack_agent.log"
	ModulesServName      = "UTMStackModulesLogsCollector"
	WinServName          = "UTMStackWindowsLogsCollector"
	ModulesLockName      = "utmstack_modules_collector.lock"
	WinLockName          = "utmstack_windows_collector.lock"
	RedlineLockName      = "utmstack_redline.lock"
	RedlineServName      = "UTMStackRedline"
	CollectorFileName    = "log-collector-configuration.json"
	CollectorFileNameOld = "log-collector-config.json"
	UUIDFileName         = "uuid.yml"
	MESSAGE_HEADER       = "utm_stack_agent_ds"
	BatchToSend          = 5
	PortRangeMin         = "7000"
	PortRangeMax         = "9000"
)

type LogType string

const (
	LogTypeWindowsAgent        LogType = "beats_windows_agent"
	LogTypeSyslog              LogType = "syslog"
	LogTypeVmware              LogType = "vmware"
	LogTypeLinuxAgent          LogType = "beats_linux_agent"
	LogTypeEset                LogType = "antivirus_eset"
	LogTypeKaspersky           LogType = "antivirus_kaspersky"
	LogTypeTraefikModule       LogType = "beats_traefik_module"
	LogTypeMongodbModule       LogType = "beats_mongodb_module"
	LogTypeMysqlModule         LogType = "beats_mysql_module"
	LogTypePostgresqlModule    LogType = "beats_postgresql_module"
	LogTypeRedisModule         LogType = "beats_redis_module"
	LogTypeElasticsearchModule LogType = "beats_elasticsearch_module"
	LogTypeKafkaModule         LogType = "beats_kafka_module"
	LogTypeKibanaModule        LogType = "beats_kibana_module"
	LogTypeLogstashModule      LogType = "beats_logstash_module"
	LogTypeCiscoAsa            LogType = "cisco_asa"
	LogTypeCiscoMeraki         LogType = "cisco_meraki"
	LogTypeFortinet            LogType = "firewall_fortinet"
	LogTypePaloalto            LogType = "firewall_paloalto"
	LogTypeMikrotik            LogType = "firewall_mikrotik"
	LogTypeCiscoFirepower      LogType = "cisco_firepower"
	LogTypeSophosXG            LogType = "firewall_sophos"
	LogTypeCiscoSwitch         LogType = "cisco_switch"
	LogTypeSonicwall           LogType = "firewall_sonicwall"
	LogTypeNatsModule          LogType = "beats_nats_module"
	LogTypeDeceptivebytes      LogType = "antivirus_deceptivebytes"
	LogTypeOsqueryModule       LogType = "beats_osquery_module"
	LogTypeLinuxAuditdModule   LogType = "beats_auditd_module"
	LogTypeHaproxyModule       LogType = "beats_haproxy_module"
	LogTypeNginxModule         LogType = "beats_nginx_module"
	LogTypeIisModule           LogType = "beats_iis_module"
	LogTypeApacheModule        LogType = "beats_apache_module"
	LogTypeSentinelOne         LogType = "antivirus_sentinel_one"
	LogTypeCiscoGeneric        LogType = "cisco"
	LogTypeMacOs               LogType = "macos_logs"
	LogTypeGeneric             LogType = "generic"
	LogTypeNetflow             LogType = "netflow"
	LogTypeAix                 LogType = "ibm_aix"
	LogTypePfsense             LogType = "firewall_pfsense"
	LogTypeFortiweb            LogType = "firewall_fortiweb"
)

type ProtoPort struct {
	UDP string
	TCP string
}

var (
	ProtoPorts = map[LogType]ProtoPort{
		LogTypeSyslog:         {UDP: "7014", TCP: "7014"},
		LogTypeVmware:         {UDP: "7002", TCP: "7002"},
		LogTypeEset:           {UDP: "7003", TCP: "7003"},
		LogTypeKaspersky:      {UDP: "7004", TCP: "7004"},
		LogTypeCiscoGeneric:   {UDP: "514", TCP: "1470"},
		LogTypeFortinet:       {UDP: "7005", TCP: "7005"},
		LogTypePaloalto:       {UDP: "7006", TCP: "7006"},
		LogTypeMikrotik:       {UDP: "7007", TCP: "7007"},
		LogTypeSophosXG:       {UDP: "7008", TCP: "7008"},
		LogTypeSonicwall:      {UDP: "7009", TCP: "7009"},
		LogTypeDeceptivebytes: {UDP: "7010", TCP: "7010"},
		LogTypeSentinelOne:    {UDP: "7012", TCP: "7012"},
		LogTypeMacOs:          {UDP: "7015", TCP: "7015"},
		LogTypeAix:            {UDP: "7016", TCP: "7016"},
		LogTypePfsense:        {UDP: "7017", TCP: "7017"},
		LogTypeFortiweb:       {UDP: "7018", TCP: "7018"},
		LogTypeNetflow:        {UDP: "2055", TCP: ""},
	}

	ProhibitedPortsChange = []LogType{LogTypeCiscoGeneric, LogTypeNetflow}
)

func GetCertPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

func GetKeyPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.key")
}

func GetCaPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, "certs", "ca.crt")
}

func GetCollectorConfigPath() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, CollectorFileName)
}

func GetCollectorConfigPathOld() string {
	path, _ := utils.GetMyPath()
	return filepath.Join(path, CollectorFileNameOld)
}

func GetAgentBin() string {
	var bin string
	switch runtime.GOOS {
	case "windows":
		bin = "utmstack_agent_service.exe"
	case "linux":
		bin = "utmstack_agent_service"
	}
	return bin
}

func GetMessageFormated(host string, msg string) string {
	return "[" + MESSAGE_HEADER + "=" + host + "]-" + msg
}

func ValidateModuleType(typ LogType) string {
	switch typ {
	case LogTypeSyslog, LogTypeVmware, LogTypeEset, LogTypeKaspersky, LogTypeFortinet, LogTypePaloalto,
		LogTypeMikrotik, LogTypeSophosXG, LogTypeSonicwall, LogTypeSentinelOne, LogTypeCiscoGeneric, LogTypeMacOs,
		LogTypeDeceptivebytes, LogTypeAix, LogTypePfsense, LogTypeFortiweb:
		return "syslog"
	case LogTypeNetflow:
		return "netflow"
	case LogTypeWindowsAgent, LogTypeLinuxAgent, LogTypeTraefikModule, LogTypeMongodbModule, LogTypeMysqlModule, LogTypePostgresqlModule,
		LogTypeRedisModule, LogTypeElasticsearchModule, LogTypeKafkaModule, LogTypeKibanaModule, LogTypeLogstashModule, LogTypeNatsModule,
		LogTypeOsqueryModule, LogTypeLinuxAuditdModule, LogTypeHaproxyModule, LogTypeNginxModule, LogTypeIisModule, LogTypeApacheModule:
		return "beats"
	default:
		return "nil"
	}
}
