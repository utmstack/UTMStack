package config

import (
	"os"
	"path/filepath"
	"time"
)

func KeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/UpdateAgent",
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

var (
	PanelConnectionKeyUrl     = "%s/api/authenticateFederationServiceManager"
	CheckEvery                = 5 * time.Minute
	CertPath                  = "/cert/utm.crt"
	CertKeyPath               = "/cert/utm.key"
	UpdatesFolder             = "/updates"
	UpdatesVersionsPath       = filepath.Join(UpdatesFolder, "version.json")
	UpdatesDependenciesFolder = "/dependencies"
	InternalKey               = os.Getenv("INTERNAL_KEY")
	PanelServiceName          = os.Getenv("PANEL_SERV_NAME")
	LogLevel                  = os.Getenv("LOG_LEVEL")
	EncryptionKey             = os.Getenv("ENCRYPTION_KEY")
	UTMHost                   = os.Getenv("UTM_HOST")
	DBHost                    = os.Getenv("DB_HOST")
	DBPort                    = os.Getenv("DB_PORT")
	DBUser                    = os.Getenv("DB_USER")
	DBPassword                = os.Getenv("DB_PASSWORD")
	DBName                    = os.Getenv("DB_NAME")
)
