package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPartialPolicyBlocksPreserveDisableAndDefaults(t *testing.T) {
	previous := ConfigFile
	ConfigFile = filepath.Join(t.TempDir(), "edr.json")
	defer func() { ConfigFile = previous }()
	for _, raw := range []string{`{"ransomware":{"enabled":false},"blocklist":{"enabled":false}}`, `{"ransomware":{"enabled":false},"blocklist":{"enforce":false}}`} {
		if err := os.WriteFile(ConfigFile, []byte(raw), 0600); err != nil {
			t.Fatal(err)
		}
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.Ransomware.Enabled || cfg.Ransomware.ResponseMode != Default().Ransomware.ResponseMode || cfg.Ransomware.SuspendThreshold == 0 {
			t.Fatalf("ransomware overlay %+v", cfg.Ransomware)
		}
		if strings.Contains(raw, `"enforce"`) {
			if cfg.Blocklist.Enforce || !cfg.Blocklist.Enabled {
				t.Fatalf("enforcement overlay %+v", cfg.Blocklist)
			}
		} else if cfg.Blocklist.Enabled {
			t.Fatal("disabled blocklist enabled")
		}
		if cfg.Blocklist.RefreshHours != Default().Blocklist.RefreshHours {
			t.Fatalf("missing defaults %+v", cfg.Blocklist)
		}
	}
	if err := os.WriteFile(ConfigFile, []byte(`{"padding":"`+strings.Repeat("x", 1<<20)+`"}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(); err == nil {
		t.Fatal("unbounded configuration accepted")
	}
}

func TestSaveCannotProduceAnUnreadableOversizedConfiguration(t *testing.T) {
	previousFile, previousDir := ConfigFile, InstallDir
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	defer func() { ConfigFile = previousFile; InstallDir = previousDir }()
	cfg := Default()
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Allowlist.Paths = []string{strings.Repeat("x", 1<<20)}
	if err = Save(cfg); err == nil {
		t.Fatal("saved configuration cannot be loaded")
	}
	after, err := os.ReadFile(ConfigFile)
	if err != nil || string(after) != string(before) {
		t.Fatal("rejected save changed previous configuration")
	}
}
