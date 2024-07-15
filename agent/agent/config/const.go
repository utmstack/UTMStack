package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/utmstack/UTMStack/agent/agent/utils"
)

const REPLACE_KEY string = ""

const (
	AGENTMANAGERPORT      string = "9000"
	CORRELATIONLOGSPORT   string = "50051"
	MaxConnectionTime            = 120 * time.Second
	InitialReconnectDelay        = 10 * time.Second
	MaxReconnectDelay            = 120 * time.Second

	SERV_NAME            = "UTMStackAgent"
	SERV_LOG             = "utmstack_agent.log"
	ModulesServName      = "UTMStackModulesLogsCollector"
	WinServName          = "UTMStackWindowsLogsCollector"
	ModulesLockName      = "utmstack_modules_collector.lock"
	WinLockName          = "utmstack_windows_collector.lock"
	CollectorFileName    = "log-collector-configuration.json"
	CollectorFileNameOld = "log-collector-config.json"
	UUIDFileName         = "uuid.yml"
	DEPEND_ZIP_LABEL     = "depend_zip"
	DEPEND_SERVICE_LABEL = "service"
	BatchToSend          = 5
	PortRangeMin         = "7000"
	PortRangeMax         = "9000"
	RetentionConfigFile  = "retention.json"
	LogsDBFile           = "logs.db"
)

type LogType struct {
	DataType string
	Config   string
}

var (
	LogTypeWindowsAgent        = LogType{DataType: "wineventlog", Config: "beats_windows_agent"}
	LogTypeSyslog              = LogType{DataType: "syslog", Config: "syslog"}
	LogTypeVmware              = LogType{DataType: "vmware-esxi", Config: "vmware"}
	LogTypeLinuxAgent          = LogType{DataType: "linux", Config: "beats_linux_agent"}
	LogTypeEset                = LogType{DataType: "antivirus-esmc-eset", Config: "antivirus_eset"}
	LogTypeKaspersky           = LogType{DataType: "antivirus-kaspersky", Config: "antivirus_kaspersky"}
	LogTypeTraefikModule       = LogType{DataType: "traefik", Config: "beats_traefik_module"}
	LogTypeMongodbModule       = LogType{DataType: "mongodb", Config: "beats_mongodb_module"}
	LogTypeMysqlModule         = LogType{DataType: "mysql", Config: "beats_mysql_module"}
	LogTypePostgresqlModule    = LogType{DataType: "postgresql", Config: "beats_postgresql_module"}
	LogTypeRedisModule         = LogType{DataType: "redis", Config: "beats_redis_module"}
	LogTypeElasticsearchModule = LogType{DataType: "elasticsearch", Config: "beats_elasticsearch_module"}
	LogTypeKafkaModule         = LogType{DataType: "kafka", Config: "beats_kafka_module"}
	LogTypeKibanaModule        = LogType{DataType: "kibana", Config: "beats_kibana_module"}
	LogTypeLogstashModule      = LogType{DataType: "logstash", Config: "beats_logstash_module"}
	LogTypeCiscoAsa            = LogType{DataType: "firewall-cisco-asa", Config: "cisco_asa"}
	LogTypeCiscoMeraki         = LogType{DataType: "firewall-meraki", Config: "cisco_meraki"}
	LogTypeFortinet            = LogType{DataType: "firewall-fortigate-traffic", Config: "firewall_fortinet"}
	LogTypePaloalto            = LogType{DataType: "firewall-paloalto", Config: "firewall_paloalto"}
	LogTypeMikrotik            = LogType{DataType: "firewall-mikrotik", Config: "firewall_mikrotik"}
	LogTypeCiscoFirepower      = LogType{DataType: "firewall-cisco-firepower", Config: "cisco_firepower"}
	LogTypeSophosXG            = LogType{DataType: "firewall-sophos-xg", Config: "firewall_sophos"}
	LogTypeCiscoSwitch         = LogType{DataType: "cisco-switch", Config: "cisco_switch"}
	LogTypeSonicwall           = LogType{DataType: "firewall-sonicwall", Config: "firewall_sonicwall"}
	LogTypeNatsModule          = LogType{DataType: "nats", Config: "beats_nats_module"}
	LogTypeDeceptivebytes      = LogType{DataType: "deceptive-bytes", Config: "antivirus_deceptivebytes"}
	LogTypeOsqueryModule       = LogType{DataType: "osquery", Config: "beats_osquery_module"}
	LogTypeLinuxAuditdModule   = LogType{DataType: "auditd", Config: "beats_auditd_module"}
	LogTypeHaproxyModule       = LogType{DataType: "haproxy", Config: "beats_haproxy_module"}
	LogTypeNginxModule         = LogType{DataType: "nginx", Config: "beats_nginx_module"}
	LogTypeIisModule           = LogType{DataType: "iis", Config: "beats_iis_module"}
	LogTypeApacheModule        = LogType{DataType: "apache", Config: "beats_apache_module"}
	LogTypeSentinelOne         = LogType{DataType: "antivirus-sentinel-one", Config: "antivirus_sentinel_one"}
	LogTypeCiscoGeneric        = LogType{DataType: "", Config: "cisco"}
	LogTypeMacOs               = LogType{DataType: "macos", Config: "macos_logs"}
	LogTypeGeneric             = LogType{DataType: "generic", Config: "generic"}
	LogTypeNetflow             = LogType{DataType: "netflow", Config: "netflow"}
	LogTypeAix                 = LogType{DataType: "ibm-aix", Config: "ibm_aix"}
	LogTypePfsense             = LogType{DataType: "firewall-pfsense", Config: "firewall_pfsense"}
	LogTypeFortiweb            = LogType{DataType: "firewall-fortiweb", Config: "firewall_fortiweb"}
)

type ProtoPort struct {
	UDP string
	TCP string
}

var (
	ProtoPorts = map[string]ProtoPort{
		LogTypeSyslog.Config:         {UDP: "7014", TCP: "7014"},
		LogTypeVmware.Config:         {UDP: "7002", TCP: "7002"},
		LogTypeEset.Config:           {UDP: "7003", TCP: "7003"},
		LogTypeKaspersky.Config:      {UDP: "7004", TCP: "7004"},
		LogTypeCiscoGeneric.Config:   {UDP: "514", TCP: "1470"},
		LogTypeFortinet.Config:       {UDP: "7005", TCP: "7005"},
		LogTypePaloalto.Config:       {UDP: "7006", TCP: "7006"},
		LogTypeMikrotik.Config:       {UDP: "7007", TCP: "7007"},
		LogTypeSophosXG.Config:       {UDP: "7008", TCP: "7008"},
		LogTypeSonicwall.Config:      {UDP: "7009", TCP: "7009"},
		LogTypeDeceptivebytes.Config: {UDP: "7010", TCP: "7010"},
		LogTypeSentinelOne.Config:    {UDP: "7012", TCP: "7012"},
		LogTypeMacOs.Config:          {UDP: "7015", TCP: "7015"},
		LogTypeAix.Config:            {UDP: "7016", TCP: "7016"},
		LogTypePfsense.Config:        {UDP: "7017", TCP: "7017"},
		LogTypeFortiweb.Config:       {UDP: "7018", TCP: "7018"},
		LogTypeNetflow.Config:        {UDP: "2055", TCP: ""},
	}

	ProhibitedPortsChange = []LogType{LogTypeCiscoGeneric, LogTypeNetflow}
)

func GetCertPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.crt")
}

func GetKeyPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "utm.key")
}

func GetCaPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "certs", "ca.crt")
}

func GetCollectorConfigPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, CollectorFileName)
}

func GetCollectorConfigPathOld() string {
	path := utils.GetMyPath()
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

func ValidateModuleType(typ string) string {
	switch typ {
	case LogTypeSyslog.Config, LogTypeVmware.Config, LogTypeEset.Config, LogTypeKaspersky.Config, LogTypeFortinet.Config, LogTypePaloalto.Config,
		LogTypeMikrotik.Config, LogTypeSophosXG.Config, LogTypeSonicwall.Config, LogTypeSentinelOne.Config, LogTypeCiscoGeneric.Config, LogTypeMacOs.Config,
		LogTypeDeceptivebytes.Config, LogTypeAix.Config, LogTypePfsense.Config, LogTypeFortiweb.Config:
		return "syslog"
	case LogTypeNetflow.Config:
		return "netflow"
	case LogTypeWindowsAgent.Config, LogTypeLinuxAgent.Config, LogTypeTraefikModule.Config, LogTypeMongodbModule.Config, LogTypeMysqlModule.Config, LogTypePostgresqlModule.Config,
		LogTypeRedisModule.Config, LogTypeElasticsearchModule.Config, LogTypeKafkaModule.Config, LogTypeKibanaModule.Config, LogTypeLogstashModule.Config, LogTypeNatsModule.Config,
		LogTypeOsqueryModule.Config, LogTypeLinuxAuditdModule.Config, LogTypeHaproxyModule.Config, LogTypeNginxModule.Config, LogTypeIisModule.Config, LogTypeApacheModule.Config:
		return "beats"
	default:
		return "nil"
	}
}

func GetDependenPaths() []string {
	path := utils.GetMyPath()
	self := "utmstack_updater_self"
	if runtime.GOOS == "windows" {
		self += ".exe"
	}
	return []string{
		filepath.Join(path, "beats"),
		filepath.Join(path, "templates"),
		filepath.Join(path, self),
	}
}

func GetDownloadFilePath(typ string, subfix string) string {
	path := utils.GetMyPath()
	switch typ {
	case DEPEND_SERVICE_LABEL:
		switch runtime.GOOS {
		case "windows":
			return filepath.Join(path, fmt.Sprintf("utmstack_agent_service%s.exe", subfix))
		case "linux":
			return filepath.Join(path, fmt.Sprintf("utmstack_agent_service%s", subfix))
		}
	case DEPEND_ZIP_LABEL:
		return filepath.Join(path, "dependencies.zip")
	}
	return ""
}

func GetVersionPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "version.json")
}

func GetVersionOldPath() string {
	path := utils.GetMyPath()
	return filepath.Join(path, "versions.json")
}

func GetSelfUpdaterPath() string {
	path := utils.GetMyPath()
	if runtime.GOOS == "windows" {
		return filepath.Join(path, "utmstack_updater_self.exe")
	}
	return filepath.Join(path, "utmstack_updater_self")
}

func GetDependenciesServices() []string {
	services := []string{"UTMStackModulesLogsCollector"}
	if runtime.GOOS == "windows" {
		services = append(services, "UTMStackWindowsLogsCollector")
	}
	return services
}
