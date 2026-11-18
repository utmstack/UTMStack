package service

import (
	"errors"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

type testHealth struct{ err error }

func (h testHealth) Health() error { return h.err }

func TestPlatformStatusDistinguishesRequestedUnsupportedAndObserved(t *testing.T) {
	cfg := config.Default()
	cfg.Enabled = true
	health := map[string]interface{ Health() error }{"file_watcher": testHealth{}, "process_guard": testHealth{errors.New("proc permission denied")}}
	linux := platformSensors(cfg, "linux", health)
	for _, name := range []string{"amsi", "network_enforcement", "ransomware"} {
		if linux[name].Supported || linux[name].State != "unsupported" {
			t.Fatalf("Linux %s: %+v", name, linux[name])
		}
	}
	if linux["file_watcher"].State != "active" || linux["process_guard"].State != "degraded" || linux["process_guard"].Error == "" {
		t.Fatalf("health %+v", linux)
	}
	win := platformSensors(cfg, "windows", nil)
	if !win["amsi"].Supported || win["amsi"].State != "runtime_unverified" {
		t.Fatalf("Windows %+v", win)
	}
	cfg.Enabled = false
	for name, st := range platformSensors(cfg, "linux", health) {
		if st.State != "disabled" {
			t.Errorf("disabled %s: %+v", name, st)
		}
	}
}

func TestFanotifyPrerequisiteAndCanaryCoverageCannotReportActive(t *testing.T) {
	cfg := config.Default()
	cfg.Enabled = true
	cfg.LinuxProtection.Fanotify.Enabled = true
	cfg.LinuxProtection.Fanotify.Mode = "permission_exec"
	off := false
	cfg.Sensors.FileWatcher = &off
	states := platformSensors(cfg, "linux", nil)
	for _, name := range []string{"file_prevention", "file_access_monitor"} {
		if !states[name].Requested || states[name].State != "degraded" || states[name].Error == "" {
			t.Fatalf("%s %+v", name, states[name])
		}
	}
	cfg.Ransomware.Enabled = true
	states = platformSensors(cfg, "linux", map[string]interface{ Health() error }{"ransomware": combinedHealth{testHealth{}, fixedHealth{errors.New("no verified canaries")}}})
	if states["ransomware"].State != "degraded" {
		t.Fatal(states["ransomware"])
	}
}
