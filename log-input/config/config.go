package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ListenAddr string

	// HealthAddr is plain HTTP and separate from ListenAddr: the gRPC port is
	// behind the auth interceptors, so a probe there would need credentials.
	HealthAddr string

	CertFile string
	KeyFile  string

	NATSURL string

	// Shards bounds how far the consumer side can be partitioned later.
	// Lowering it strands messages already published to the higher subjects.
	Shards int

	RedisAddr     string
	RedisPassword string
	RedisDB       int

	AgentManager string
	Backend      string
	InternalKey  string

	DefaultTenant string

	// AuthTTL bounds how long a revoked key keeps being accepted.
	AuthTTL time.Duration
}

const (
	defaultListenAddr   = "0.0.0.0:50051"
	defaultHealthAddr   = "0.0.0.0:8080"
	defaultShards       = 16
	defaultTenant       = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	defaultAuthTTL      = 5 * time.Minute
	defaultCertsFolder  = "/cert"
	utmCertFileName     = "utm.crt"
	utmCertFileKeyName  = "utm.key"
	maxReasonableShards = 1024
	minAuthTTL          = time.Second
)

func Load() (*Config, error) {
	certs := envOr("CERTS_FOLDER", defaultCertsFolder)

	c := &Config{
		ListenAddr:    envOr("LISTEN_ADDR", defaultListenAddr),
		HealthAddr:    envOr("HEALTH_ADDR", defaultHealthAddr),
		CertFile:      certs + "/" + utmCertFileName,
		KeyFile:       certs + "/" + utmCertFileKeyName,
		NATSURL:       os.Getenv("NATS_URL"),
		Shards:        envInt("SHARDS", defaultShards),
		RedisAddr:     os.Getenv("REDIS_ADDR"),
		RedisPassword: os.Getenv("REDIS_PASSWORD"),
		RedisDB:       envInt("REDIS_DB", 0),
		AgentManager:  os.Getenv("AGENT_MANAGER"),
		Backend:       os.Getenv("BACKEND"),
		InternalKey:   os.Getenv("INTERNAL_KEY"),
		DefaultTenant: envOr("DEFAULT_TENANT", defaultTenant),
		AuthTTL:       envDuration("AUTH_TTL", defaultAuthTTL),
	}

	// Refused rather than defaulted: accepting a log with nowhere to put it
	// loses it.
	if c.NATSURL == "" {
		return nil, fmt.Errorf("NATS_URL is required")
	}
	if c.RedisAddr == "" {
		return nil, fmt.Errorf("REDIS_ADDR is required")
	}
	if c.AgentManager == "" {
		return nil, fmt.Errorf("AGENT_MANAGER is required")
	}
	if c.InternalKey == "" {
		return nil, fmt.Errorf("INTERNAL_KEY is required")
	}
	if c.Shards < 1 || c.Shards > maxReasonableShards {
		return nil, fmt.Errorf("SHARDS must be between 1 and %d, got %d", maxReasonableShards, c.Shards)
	}
	if c.AuthTTL < minAuthTTL {
		return nil, fmt.Errorf("AUTH_TTL must be at least %s", minAuthTTL)
	}

	return c, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
