package config

import (
	"os"
	"strings"
	"testing"
)

func TestDefaultHasSaneValues(t *testing.T) {
	d := Default()
	if d.FailMode != "open" {
		t.Fatalf("FailMode = %q, want open", d.FailMode)
	}
	if d.SuspendOnLaunch {
		t.Fatal("SuspendOnLaunch should default false (suspending every launch is unsafe without a kernel driver)")
	}
	if d.SuspendTimeoutMs <= 0 {
		t.Fatalf("SuspendTimeoutMs = %d, want a positive bound", d.SuspendTimeoutMs)
	}
	if d.ClamdAddr != "127.0.0.1:3310" {
		t.Fatalf("ClamdAddr = %q", d.ClamdAddr)
	}
	if d.ScanConcurrency < 1 {
		t.Fatalf("ScanConcurrency = %d, want >=1", d.ScanConcurrency)
	}
}

func TestLoadReturnsDefaultWhenNoFile(t *testing.T) {
	// Point ConfigFile at a non-existent temp path.
	orig := ConfigFile
	ConfigFile = t.TempDir() + "/nope.json"
	defer func() { ConfigFile = orig }()
	c, err := Load()
	if err != nil {
		t.Fatalf("Load err: %v", err)
	}
	if c.FailMode != "open" {
		t.Fatalf("expected default FailMode, got %q", c.FailMode)
	}
}

func TestLoadMigratesLegacyWhitelistsIntoAllowlist(t *testing.T) {
	orig := ConfigFile
	ConfigFile = t.TempDir() + "/edr.json"
	defer func() { ConfigFile = orig }()

	// A pre-existing edr.json using the OLD scattered keys.
	legacy := `{
      "enabled": true,
      "exclusions": ["C:\\builds"],
      "trusted_processes": ["C:\\Program Files\\Veeam"],
      "ransomware": {"enabled": true, "response_mode": "kill", "command_allowlist": ["/oldest"]}
    }`
	if err := os.WriteFile(ConfigFile, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Allowlist.Paths) != 1 || c.Allowlist.Paths[0] != `C:\builds` {
		t.Fatalf("legacy exclusions not migrated: %+v", c.Allowlist.Paths)
	}
	if len(c.Allowlist.Processes) != 1 || c.Allowlist.Processes[0] != `C:\Program Files\Veeam` {
		t.Fatalf("legacy trusted_processes not migrated: %+v", c.Allowlist.Processes)
	}
	if len(c.Allowlist.Commands) != 1 || c.Allowlist.Commands[0] != "/oldest" {
		t.Fatalf("legacy command_allowlist not migrated: %+v", c.Allowlist.Commands)
	}
	// The deprecated fields are cleared post-migration so Save writes the unified form.
	if c.Exclusions != nil || c.TrustedProcesses != nil || c.Ransomware.CommandAllowlist != nil {
		t.Fatalf("deprecated fields not cleared after migration")
	}

	// Save then reload — the unified allowlist must survive a round-trip and the
	// old keys must be gone from disk.
	if err := Save(c); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(ConfigFile)
	if strings.Contains(string(raw), "trusted_processes") || strings.Contains(string(raw), "command_allowlist") {
		t.Fatalf("Save re-wrote deprecated keys:\n%s", raw)
	}
	c2, _ := Load()
	if len(c2.Allowlist.Processes) != 1 || len(c2.Allowlist.Commands) != 1 {
		t.Fatalf("unified allowlist not preserved across round-trip: %+v", c2.Allowlist)
	}
}

func TestSensorsDefaultOnAndToggle(t *testing.T) {
	d := Default()
	if !d.Sensors.FileWatcherOn() || !d.Sensors.ProcessGuardOn() || !d.Sensors.AMSIOn() || !d.Sensors.BehavioralOn() {
		t.Fatal("all sensors must default ON (nil pointers)")
	}
	off := false
	s := Sensors{Behavioral: &off}
	if s.BehavioralOn() {
		t.Fatal("explicit false must turn a sensor off")
	}
	if !s.FileWatcherOn() {
		t.Fatal("an unset sibling sensor must stay on")
	}
}
