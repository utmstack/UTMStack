package config

import (
	"os"
	"strconv"
)

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBoolOrDefault(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envIntOrDefault(key string, fallback int) int {
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

var (
	RecoveryDir               = envOrDefault("RECOVERY_DIR", "/app/recoveries/")
	RecoveryDispatchEnabled   = envBoolOrDefault("RECOVERY_DISPATCH_ENABLED", true)
	RecoveryGlobalConcurrency = envIntOrDefault("RECOVERY_GLOBAL_CONCURRENCY", 50)
)
