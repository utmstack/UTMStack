// edr/internal/daemon/daemon_test.go
package daemon

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvDefaults(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{"INTERNAL_KEY": "ik"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MirrorDir != "/mirror" || cfg.SigEvery != time.Hour || cfg.FeedEvery != 6*time.Hour {
		t.Fatalf("defaults = %+v", cfg)
	}
	if cfg.HTTPPort != "9002" {
		t.Fatalf("HTTPPort = %q, want 9002", cfg.HTTPPort)
	}
	if len(cfg.Levels) != 1 || cfg.Levels[0] != 1 {
		t.Fatalf("levels = %v", cfg.Levels)
	}
	if len(cfg.Names) != 3 || cfg.Names[0] != "ip" {
		t.Fatalf("names = %v", cfg.Names)
	}
	if cfg.UTMHost != "http://backend:8080" || cfg.TWURL == "" {
		t.Fatalf("hosts = %+v", cfg)
	}
}

func TestFromEnvOverridesAndValidation(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"EDR_MIRROR_DIR": "/data", "EDR_SIG_SYNC_MINUTES": "30",
		"EDR_FEED_SYNC_HOURS": "2", "EDR_FEED_LEVELS": "1,2",
		"EDR_FEED_NAMES": "ip", "TW_API_KEY": "k", "TW_API_SECRET": "s",
		"EDR_HTTP_PORT": "8080",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MirrorDir != "/data" || cfg.SigEvery != 30*time.Minute || cfg.FeedEvery != 2*time.Hour {
		t.Fatalf("overrides = %+v", cfg)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("HTTPPort override = %q, want 8080", cfg.HTTPPort)
	}
	if len(cfg.Levels) != 2 || cfg.Levels[1] != 2 || len(cfg.Names) != 1 {
		t.Fatalf("lists = %+v", cfg)
	}

	// Neither INTERNAL_KEY nor a TW override → error.
	if _, err := FromEnv(env(map[string]string{})); err == nil {
		t.Fatal("expected error when no credential source is configured")
	}
}
