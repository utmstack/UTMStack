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

		internalKey:   env.String("INTERNAL_KEY", "", false),
		encryptionKey: env.String("ENCRYPTION_KEY", "", false),

		grpcAgentManagerHost: env.String("GRPC_AGENT_MANAGER_HOST", "agentmanager", false),
		grpcAgentManagerPort: env.Int("GRPC_AGENT_MANAGER_PORT", 9000, false),

		eventProcessorHost: env.String("EVENT_PROCESSOR_HOST", "event-processor-manager", false),
		eventProcessorPort: env.Int("EVENT_PROCESSOR_PORT", 9002, false),

		socAIBaseURL: env.String("SOC_AI_BASE_URL", "", false),

		uploadDir: env.String("UPLOAD_DIR", "./uploads", false),

		tfaEnabled: env.Bool("APP_TFA_ENABLED", true),
	}
}
