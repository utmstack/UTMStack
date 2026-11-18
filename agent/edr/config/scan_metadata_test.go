package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestCollectMetadataDefaultsOffAndPersists(t *testing.T) {
	if Default().CollectMetadata {
		t.Fatal("metadata collection must default off like ClamAV")
	}
	var decoded EDRConfig
	if err := json.Unmarshal([]byte(`{"collect_metadata":true}`), &decoded); err != nil || !decoded.CollectMetadata {
		t.Fatalf("collect_metadata JSON: enabled=%v err=%v", decoded.CollectMetadata, err)
	}
	oldInstallDir, oldConfigFile := InstallDir, ConfigFile
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	t.Cleanup(func() { InstallDir, ConfigFile = oldInstallDir, oldConfigFile })
	cfg := Default()
	cfg.CollectMetadata = true
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil || !loaded.CollectMetadata {
		t.Fatalf("collect_metadata Save/Load: enabled=%v err=%v", loaded.CollectMetadata, err)
	}
}
