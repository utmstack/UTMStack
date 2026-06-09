package main

import "github.com/utmstack/utmstack/backend/pkg/env"

type config struct {
	// Server
	appPort    int
	devMode    bool
	serverName string
	jwtIssuer  string

	// JWT
	jwtSecret string

	// Postgres
	dbHost string
	dbPort int
	dbName string
	dbUser string
	dbPass string

	// Elasticsearch
	esHost     string
	esPort     int
	esUser     string
	esPassword string

	// OpenSearch disk-space guard
	diskGuardEnabled     bool
	diskWarnPercent      float64
	diskDeletePercent    float64
	diskGuardIntervalSec int

	// Secrets
	internalKey   string
	encryptionKey string

	// gRPC: agent manager
	grpcAgentManagerHost string
	grpcAgentManagerPort int

	// gRPC: event processor
	eventProcessorHost string
	eventProcessorPort int

	// SOC AI
	socAIBaseURL string

	// Uploads (avatars, etc.)
	uploadDir string

	// Features
	tfaEnabled bool

	// MCP (Model Context Protocol) — exposes the backend usecases to MCP-aware
	// AI clients at /api/v1/mcp. Auth/permissions/audit reuse the REST stack.
	mcpEnabled bool
	mcpVersion string
}

func loadConfig() *config {
	return &config{
		appPort:    env.Int("APP_PORT", 8080, false),
		devMode:    env.Bool("DEV_MODE", false),
		serverName: env.String("SERVER_NAME", "v11dev", false),
		jwtIssuer:  env.String("JWT_ISSUER", "", false),

		jwtSecret: env.String("JWT_SECRET", "", false),

		dbHost: env.String("DB_HOST", "localhost", false),
		dbPort: env.Int("DB_PORT", 5432, false),
		dbName: env.String("DB_NAME", "utmstack", false),
		dbUser: env.String("DB_USER", "postgres", false),
		dbPass: env.String("DB_PASS", "", false),

		esHost:     env.String("ELASTICSEARCH_HOST", "localhost", false),
		esPort:     env.Int("ELASTICSEARCH_PORT", 9200, false),
		esUser:     env.String("ELASTICSEARCH_USER", "admin", false),
		esPassword: env.String("ELASTICSEARCH_PASSWORD", "", false),

		diskGuardEnabled:     env.Bool("DISK_GUARD_ENABLED", true),
		diskWarnPercent:      float64(env.Int("DISK_WARN_PERCENT", 70, false)),
		diskDeletePercent:    float64(env.Int("DISK_DELETE_PERCENT", 85, false)),
		diskGuardIntervalSec: env.Int("DISK_GUARD_INTERVAL_SECONDS", 60, false),

		internalKey:   env.String("INTERNAL_KEY", "", false),
		encryptionKey: env.String("ENCRYPTION_KEY", "", false),

		grpcAgentManagerHost: env.String("GRPC_AGENT_MANAGER_HOST", "agentmanager", false),
		grpcAgentManagerPort: env.Int("GRPC_AGENT_MANAGER_PORT", 9000, false),

		eventProcessorHost: env.String("EVENT_PROCESSOR_HOST", "event-processor-manager", false),
		eventProcessorPort: env.Int("EVENT_PROCESSOR_PORT", 9002, false),

		socAIBaseURL: env.String("SOC_AI_BASE_URL", "", false),

		uploadDir: env.String("UPLOAD_DIR", "./uploads", false),

		tfaEnabled: env.Bool("APP_TFA_ENABLED", true),

		mcpEnabled: env.Bool("MCP_ENABLED", true),
		mcpVersion: env.String("MCP_SERVER_VERSION", "v1", false),
	}
}
