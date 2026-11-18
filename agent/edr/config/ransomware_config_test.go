package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRansomwareDefaults(t *testing.T) {
	d := Default()
	if !d.Ransomware.Enabled {
		t.Fatal("ransomware must default ON")
	}
	if d.Ransomware.ResponseMode != "suspend" {
		t.Fatalf("default response_mode = %q, want suspend", d.Ransomware.ResponseMode)
	}
	if d.Ransomware.KillThreshold != 100 || d.Ransomware.SuspendThreshold != 50 {
		t.Fatalf("thresholds = %d/%d", d.Ransomware.SuspendThreshold, d.Ransomware.KillThreshold)
	}
	if !d.Ransomware.UseETW || d.Ransomware.CanaryPerDir != 1 || d.Ransomware.DecayHalfLifeMs != 10000 {
		t.Fatalf("bad defaults: %+v", d.Ransomware)
	}
}

func TestRansomwareOverlayRoundTrip(t *testing.T) {
	dir := t.TempDir()
	old := ConfigFile
	ConfigFile = filepath.Join(dir, "edr.json")
	defer func() { ConfigFile = old }()

	on := Default()
	on.Ransomware.Enabled = true
	on.Ransomware.ResponseMode = "kill"
	on.Ransomware.CanaryDirs = []string{`D:\shared`}
	b, _ := json.Marshal(on)
	if err := os.WriteFile(ConfigFile, b, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ransomware.Enabled || got.Ransomware.ResponseMode != "kill" ||
		len(got.Ransomware.CanaryDirs) != 1 || got.Ransomware.CanaryDirs[0] != `D:\shared` {
		t.Fatalf("overlay lost ransomware block: %+v", got.Ransomware)
	}
}
