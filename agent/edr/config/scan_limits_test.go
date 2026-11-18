package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestScanResourceValidation(t *testing.T) {
	for _, field := range []struct {
		name string
		max  int
		set  func(*EDRConfig, int)
	}{
		{"scan_concurrency", MaxScanConcurrency, func(c *EDRConfig, n int) { c.ScanConcurrency = n }},
		{"max_file_size_mb", MaxFileSizeMB, func(c *EDRConfig, n int) { c.ScanLimits.MaxFileSizeMB = n }},
		{"max_scan_size_mb", MaxScanSizeMB, func(c *EDRConfig, n int) { c.ScanLimits.MaxScanSizeMB = n }},
		{"max_recursion", MaxRecursion, func(c *EDRConfig, n int) { c.ScanLimits.MaxRecursion = n }},
		{"max_files", MaxFiles, func(c *EDRConfig, n int) { c.ScanLimits.MaxFiles = n }},
		{"max_partitions", MaxPartitions, func(c *EDRConfig, n int) { c.ScanLimits.MaxPartitions = n }},
		{"max_scan_time_ms", MaxScanTimeMS, func(c *EDRConfig, n int) { c.ScanLimits.MaxScanTimeMS = n }},
	} {
		t.Run(field.name, func(t *testing.T) {
			for _, n := range []int{-1, field.max + 1} {
				cfg := Default()
				field.set(&cfg, n)
				if err := cfg.ValidateScanResources(); err == nil || !strings.Contains(err.Error(), field.name) {
					t.Errorf("value %d: %v, want validation naming %s", n, err, field.name)
				}
			}
			for _, n := range []int{0, 1, field.max} {
				cfg := Default()
				field.set(&cfg, n)
				if err := cfg.ValidateScanResources(); err != nil {
					t.Errorf("valid value %d rejected: %v", n, err)
				}
			}
		})
	}
	cfg := Default()
	cfg.ScanLimits = ScanLimits{MaxFileSizeMB: 100, MaxScanSizeMB: 50}
	if err := cfg.ValidateScanResources(); err == nil {
		t.Fatal("explicit file budget exceeding explicit total was accepted")
	}
	cfg.ScanLimits = ScanLimits{}
	cfg.TierOverride = "sever"
	if err := cfg.ValidateScanResources(); err == nil {
		t.Fatal("invalid tier accepted")
	}
}

func TestScanLimitsRoundTripAndInvalidEdits(t *testing.T) {
	beforeFile, beforeInstall := ConfigFile, InstallDir
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	defer func() { ConfigFile, InstallDir = beforeFile, beforeInstall }()
	cfg := Default()
	cfg.ScanConcurrency = 1
	cfg.ScanLimits = ScanLimits{MaxFileSizeMB: 5, MaxScanSizeMB: 12, MaxRecursion: 3, MaxFiles: 20, MaxPartitions: 2, MaxScanTimeMS: 750}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil || loaded.ScanLimits != cfg.ScanLimits || loaded.ScanConcurrency != cfg.ScanConcurrency {
		t.Fatalf("round-trip: limits=%+v workers=%d err=%v", loaded.ScanLimits, loaded.ScanConcurrency, err)
	}
	original, err := os.ReadFile(ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"max_file_size_mb", "max_scan_size_mb", "max_recursion", "max_files", "max_partitions", "max_scan_time_ms"} {
		var obj map[string]any
		if err := json.Unmarshal(original, &obj); err != nil {
			t.Fatal(err)
		}
		obj["scan_limits"].(map[string]any)[key] = -1
		data, _ := json.Marshal(obj)
		if err := os.WriteFile(ConfigFile, data, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(); err == nil || !strings.Contains(err.Error(), key) {
			t.Errorf("hand-edited %s silently accepted: %v", key, err)
		}
	}
	if err := os.WriteFile(ConfigFile, original, 0600); err != nil {
		t.Fatal(err)
	}
	invalid := cfg
	invalid.ScanLimits.MaxFiles = MaxFiles + 1
	if err := Save(invalid); err == nil {
		t.Fatal("Save accepted an invalid budget")
	}
	after, err := os.ReadFile(ConfigFile)
	if err != nil || !reflect.DeepEqual(original, after) {
		t.Fatalf("invalid Save changed the existing configuration: %v", err)
	}
}

func TestScanLimitsRejectMisspelledKnobs(t *testing.T) {
	var cfg EDRConfig
	if err := json.Unmarshal([]byte(`{"scan_limits":{"max_file_size_mb":5,"max_scan_siez_mb":10}}`), &cfg); err == nil || !strings.Contains(err.Error(), "max_scan_siez_mb") {
		t.Fatalf("misspelled resource budget silently accepted: %v", err)
	}
	if err := json.Unmarshal([]byte(`{"retired_daemon_knob":true,"scan_limits":{"max_file_size_mb":5}}`), &cfg); err != nil {
		t.Fatalf("legacy top-level configuration migration rejected: %v", err)
	}
}
