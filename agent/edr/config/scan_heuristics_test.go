package config

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestPDFNameHeuristicConfigRoundTrip(t *testing.T) {
	if !Default().ScanHeuristics.PDFObfuscatedNameOn() {
		t.Fatal("omitted heuristic must retain default-on behavior")
	}
	var cfg EDRConfig
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"pdf_obfuscated_name":false}}`), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ScanHeuristics.PDFObfuscatedNameOn() {
		t.Fatal("explicit false was ignored")
	}
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"pdf_obfuscated_names":false}}`), &cfg); err == nil {
		t.Fatal("misspelled heuristic key was accepted")
	}

	oldInstallDir, oldConfigFile := InstallDir, ConfigFile
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	t.Cleanup(func() { InstallDir, ConfigFile = oldInstallDir, oldConfigFile })
	cfg = Default()
	off := false
	cfg.ScanHeuristics.PDFObfuscatedName = &off
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ScanHeuristics.PDFObfuscatedNameOn() {
		t.Fatal("disabled heuristic did not survive Save/Load")
	}
}

func TestPartitionIntersectionHeuristicConfigRoundTrip(t *testing.T) {
	if Default().ScanHeuristics.PartitionIntersectionOn() {
		t.Fatal("omitted partition intersection alert must be off")
	}
	var cfg EDRConfig
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"partition_intersection":true}}`), &cfg); err != nil {
		t.Fatal(err)
	}
	if !cfg.ScanHeuristics.PartitionIntersectionOn() {
		t.Fatal("explicit true was ignored")
	}
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"partition_intersections":true}}`), &cfg); err == nil {
		t.Fatal("misspelled partition intersection key was accepted")
	}

	oldInstallDir, oldConfigFile := InstallDir, ConfigFile
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	t.Cleanup(func() { InstallDir, ConfigFile = oldInstallDir, oldConfigFile })
	cfg = Default()
	on := true
	cfg.ScanHeuristics.PartitionIntersection = &on
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.ScanHeuristics.PartitionIntersectionOn() {
		t.Fatal("enabled heuristic did not survive Save/Load")
	}
}

func TestBrokenExecutableHeuristicConfigRoundTrip(t *testing.T) {
	if !Default().ScanHeuristics.BrokenExecutableOn() {
		t.Fatal("omitted broken-executable setting must preserve the endpoint profile")
	}
	var cfg EDRConfig
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"broken_executable":false}}`), &cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.ScanHeuristics.BrokenExecutableOn() {
		t.Fatal("explicit false was ignored")
	}
	if err := json.Unmarshal([]byte(`{"scan_heuristics":{"broken_executables":false}}`), &cfg); err == nil {
		t.Fatal("misspelled key was accepted")
	}

	oldInstallDir, oldConfigFile := InstallDir, ConfigFile
	InstallDir = t.TempDir()
	ConfigFile = filepath.Join(InstallDir, "edr.json")
	t.Cleanup(func() { InstallDir, ConfigFile = oldInstallDir, oldConfigFile })
	cfg = Default()
	off := false
	cfg.ScanHeuristics.BrokenExecutable = &off
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ScanHeuristics.BrokenExecutableOn() {
		t.Fatal("disabled setting did not survive Save/Load")
	}
}
