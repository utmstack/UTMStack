package config

import "os"

func AgentKeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/Ping",
		"/agent.AgentService/DeleteAgent",
		"agent.AgentConfigService/AgentModuleUpdateStream",
		"agent.AgentConfigService/GetAgentConfig",
	}
}

func ConnectionKeyRoutes() []string {
	return []string{
		"/agent.AgentService/RegisterAgent",
		"/agent.AgentService/ListAgents",
		"/agent.AgentService/ListAgentCommands",
		"/agent.AgentService/GetAgentByHostname",
		"/agent.AgentService/ListAgentsWithCommands",
		"/agent.AgentService/UpdateAgentType",
		"/agent.AgentService/UpdateAgentGroup",
		"/agent.AgentService/ListAgentCommands",
	}
}

const PanelConnectionKeyUrl = "%s/api/authenticateFederationServiceManager"
const UTMSharedKeyEnv = "INTERNAL_KEY"
const UTMEncryptionKeyEnv = "ENCRYPTION_KEY"
const UTMHostEnv = "UTM_HOST"

func GetInternalKey() string {
	return os.Getenv(UTMSharedKeyEnv)
}
