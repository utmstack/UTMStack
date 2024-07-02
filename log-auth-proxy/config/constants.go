package config

import "os"

const PanelConnectionKeyEndpoint = "%s/api/federation-service/token"
const PanelPipelinesEndpoint = "%s/api/logstash-pipelines/port"
const LogstashPipelinesEndpoint = "http://%s:%s"
const PanelAPIKeyHeader = "Utm-Internal-Key"
const ProxyAPIKeyHeader = "Utm-Connection-Key"
const ProxyLogTypeHeader = "Utm-Log-Type"
const ProxySourceHeader = "Utm-Log-Source"
const UTMSharedKeyEnv = "INTERNAL_KEY"
const UTMHostEnv = "UTM_HOST"
const UTMAgentManagerHostEnv = "UTM_AGENT_MANAGER_HOST"
const UTMLogstashHostEnv = "UTM_LOGSTASH_HOST"
const UTMCertsLocationEnv = "UTM_CERTS_LOCATION"
const UTMCertFileName = "utm.crt"
const UTMCertFileKey = "utm.key"
const UTMLogSeparator = "<utm-log-separator>"

func LogstashHost() string {
	return os.Getenv(UTMLogstashHostEnv)
}

type LogType string

const (
	BeatsLinuxAgent          LogType = "beats_linux_agent"
	BeatsApacheModule        LogType = "beats_apache_module"
	BeatsApache2Module       LogType = "beats_apache2_module"
	BeatsAuditdModule        LogType = "beats_auditd_module"
	BeatsElasticsearchModule LogType = "beats_elasticsearch_module"
	BeatsKafkaModule         LogType = "beats_kafka_module"
	BeatsKibanaModule        LogType = "beats_kibana_module"
	BeatsLogstashModule      LogType = "beats_logstash_module"
	BeatsMongodbModule       LogType = "beats_mongodb_module"
	BeatsMysqlModule         LogType = "beats_mysql_module"
	BeatsNginxModule         LogType = "beats_nginx_module"
	BeatsOsqueryModule       LogType = "beats_osquery_module"
	BeatsPostgresqlModule    LogType = "beats_postgresql_module"
	BeatsRedisModule         LogType = "beats_redis_module"
	BeatsIisModule           LogType = "beats_iis_module"
	MacosLogs                LogType = "macos_logs"
	Vmware                   LogType = "vmware"
	CiscoAsa                 LogType = "cisco_asa"
	CiscoMeraki              LogType = "cisco_meraki"
	CiscoFirepower           LogType = "cisco_firepower"
	CiscoSwitch              LogType = "cisco_switch"
	AntivirusKaspersky       LogType = "antivirus_kaspersky"
	AntivirusEset            LogType = "antivirus_eset"
	AntivirusSentinelOne     LogType = "antivirus_sentinel_one"
	FirewallFortinet         LogType = "firewall_fortinet"
	FirewallSophos           LogType = "firewall_sophos"
	CloudAzure               LogType = "cloud_azure"
	CloudGoogle              LogType = "cloud_google"
	FirewallUbuntu           LogType = "firewall_ubuntu"
	FirewallMikrotik         LogType = "firewall_mikrotik"
	FirewallPaloalto         LogType = "firewall_paloalto"
	FirewallSonicwall        LogType = "firewall_sonicwall"
	AntivirusDeceptivebytes  LogType = "antivirus_deceptivebytes"
	WebHookGithub            LogType = "web_hook_github"
	AntivirusBitdefender     LogType = "antivirus_bitdefender"
	BeatsWindowsAgent        LogType = "beats_windows_agent"
	BeatsTraefikModule       LogType = "beats_traefik_module"
	BeatsNatsModule          LogType = "beats_nats_module"
	BeatsHaproxyModule       LogType = "beats_haproxy_module"
	JsonInput                LogType = "json_input"
	Syslog                   LogType = "syslog"
	Generic                  LogType = "generic"
	Netflow                  LogType = "netflow"
	Aix                      LogType = "ibm_aix"
	AS400                    LogType = "as_400"
	FirewallPfsense          LogType = "firewall_pfsense"
	FirewallFortiweb         LogType = "firewall_fortiweb"
)

type ServiceStatus string

const (
	Up   ServiceStatus = "up"
	Down ServiceStatus = "down"
)

type PipelinePortConfiguration struct {
	InputId LogType
	Port    string
}

func ValidateConnectorType(typ string) string {
	switch LogType(typ) {
	case AS400:
		return "collector"
	default:
		return "agent"
	}
}
