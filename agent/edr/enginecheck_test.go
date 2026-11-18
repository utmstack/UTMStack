package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
)

func TestEngineCheckReportsUnknownWhenReachedRules(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	rules := "Rx;Engine:81-255,Target:0;0&1;414243;0/x(?R)/\nPlain;Engine:51-255,Target:0;0;41424344\n"
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "local.ldb"), []byte(rules), 0o600); err != nil {
		t.Fatal(err)
	}
	e := engine.New(cfg)
	err := e.EnsureRunning()
	r := engineCheckReport(e, err)
	if !r.Healthy || r.UnknownWhenReached["pcre"] != 1 || len(r.RepresentationsNotProduced) == 0 {
		t.Fatalf("report = %+v", r)
	}
	b, _ := json.Marshal(r)
	for _, key := range []string{`"unknown_when_reached"`, `"representations_not_produced"`, `"` + engine.ImplementationVersion + `"`} {
		if !strings.Contains(string(b), key) {
			t.Fatalf("engine-check JSON lacks %s: %s", key, b)
		}
	}
	if strings.Contains(strings.ToLower(string(b)), "clam") {
		t.Fatalf("engine-check JSON leaks the engine name: %s", b)
	}
}

func TestEngineCheckReportsEffectiveScanBudgets(t *testing.T) {
	cfg := config.Default()
	cfg.ScanConcurrency = 1
	cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 1, MaxScanSizeMB: 2, MaxRecursion: 3, MaxFiles: 20, MaxScanTimeMS: 75}
	e := engine.New(cfg)
	r := engineCheckReport(e, engine.ErrUnavailable)
	if r.Healthy {
		t.Fatal("unavailable engine reported healthy")
	}
	data, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		t.Fatal(err)
	}
	effective, ok := obj["effective_tuning"].(map[string]any)
	if !ok {
		t.Fatalf("effective limits absent: %s", data)
	}
	for key, want := range map[string]float64{"max_threads": 1, "max_file_size_mb": 1, "max_scan_size_mb": 2, "max_recursion": 3, "max_files": 20, "max_scan_time_ms": 75} {
		if effective[key] != want {
			t.Errorf("%s = %v, want %v", key, effective[key], want)
		}
	}
	if _, misleading := effective["max_queue"]; misleading {
		t.Fatal("legacy queue estimate reported as an effective native limit")
	}
}
