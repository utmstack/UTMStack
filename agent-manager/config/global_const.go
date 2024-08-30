package config

import (
	"os"
	"time"
)

func KeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/DeleteAgent",

		"/agent.CollectorService/CollectorStream",
		"/agent.CollectorService/DeleteCollector",
		"/agent.CollectorService/GetCollectorConfig",

		"/agent.ModuleConfigService/IsModuleEnabled",

		"/agent.PingService/Ping",

		"/grpc.health.v1.Health/Check",
	}
}

func ConnectionKeyRoutes() []string {
	return []string{
		"/agent.AgentService/RegisterAgent",
		"/agent.CollectorService/RegisterCollector",
	}
}

func InternalKeyRoutes() []string {
	return []string{
		"/agent.AgentService/ListAgents",
		"/agent.AgentService/ListAgentCommands",

		"/agent.CollectorService/ListCollector",

		"/agent.PanelService/ProcessCommand",
		"/agent.PanelCollectorService/RegisterCollectorConfig",

		"/grpc.health.v1.Health/Check",
	}
}

const (
	PanelConnectionKeyUrl = "%s/api/authenticateFederationServiceManager"
	MASTERVERSIONENDPOINT = "/management/info"
	Bucket                = "https://cdn.utmstack.com/"
	VOLPATH               = "/data"
	VersionsFile          = "version.json"
	CHECK_EVERY           = 5 * time.Minute
	DependInstallerLabel  = "installer"
	DependServiceLabel    = "service"
	DependZipLabel        = "depend_zip"
	CertPath              = "/cert/utm.crt"
	CertKeyPath           = "/cert/utm.key"
)

var (
	DependOsAllows = []string{"windows", "linux"}
	DependTypes    = []string{DependInstallerLabel, DependServiceLabel, DependZipLabel}
)

func GetInternalKey() string {
	return os.Getenv("INTERNAL_KEY")
}

func GetPanelServiceName() string {
	return os.Getenv("PANEL_SERV_NAME")
}

func GetLogLevel() string {
	return os.Getenv("LOG_LEVEL")
}

func GetEncryptionKey() string {
	return os.Getenv("ENCRYPTION_KEY")
}

func GetUTMHost() string {
	return os.Getenv("UTM_HOST")
}

func GetDBHost() string {
	return os.Getenv("DB_HOST")
}

func GetDBPort() string {
	return os.Getenv("DB_PORT")
}

func GetDBUser() string {
	return os.Getenv("DB_USER")
}

func GetDBPassword() string {
	return os.Getenv("DB_PASSWORD")
}

func GetDBName() string {
	return os.Getenv("DB_NAME")
}
