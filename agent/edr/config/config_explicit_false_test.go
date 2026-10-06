package config

import (
	"os"
	"path/filepath"
	"testing"
)

// H5: the config-safety trap. An explicitly written `enabled: false` for the
// blocklist or ransomware block must be honoured after a restart. An absent
// block must still take the default (on). Three states per block, tested
// against a real file written to disk and read back through Load().

func writeCfg(t *testing.T, raw string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "edr.json")
	old := ConfigFile
	ConfigFile = path
	t.Cleanup(func() { ConfigFile = old })
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestH5BlocklistExplicitFalseHonoured(t *testing.T) {
	writeCfg(t, `{"blocklist":{"enabled":false}}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Blocklist.EnabledOn() {
		t.Fatal("explicit enabled:false must be honoured, got enabled")
	}
}

func TestH5BlocklistAbsentTakesDefault(t *testing.T) {
	writeCfg(t, `{"enabled":true}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Blocklist.EnabledOn() {
		t.Fatal("absent blocklist block should default to enabled")
	}
}

func TestH5BlocklistExplicitTrueHonoured(t *testing.T) {
	writeCfg(t, `{"blocklist":{"enabled":true}}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Blocklist.EnabledOn() {
		t.Fatal("explicit enabled:true must be honoured")
	}
}

func TestH5RansomwareExplicitFalseHonoured(t *testing.T) {
	writeCfg(t, `{"ransomware":{"enabled":false}}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got.Ransomware.EnabledOn() {
		t.Fatal("explicit enabled:false must be honoured, got enabled")
	}
}

func TestH5RansomwareAbsentTakesDefault(t *testing.T) {
	writeCfg(t, `{"enabled":true}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ransomware.EnabledOn() {
		t.Fatal("absent ransomware block should default to enabled")
	}
	if got.Ransomware.ResponseMode != "suspend" {
		t.Fatalf("absent ransomware block should keep default response_mode, got %q", got.Ransomware.ResponseMode)
	}
}

func TestH5RansomwareExplicitTrueHonoured(t *testing.T) {
	writeCfg(t, `{"ransomware":{"enabled":true}}`)
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Ransomware.EnabledOn() {
		t.Fatal("explicit enabled:true must be honoured")
	}
}

// LogEffective must report settings that differ from the default.
// It is safe to call with the global logger uninitialised (no-op).
func TestH5LogEffectiveReportsDiffs(t *testing.T) {
	d := Default()
	if n := LogEffective(d); n != 0 {
		t.Fatalf("default config should log 0 diffs, got %d", n)
	}

	c := Default()
	c.Blocklist.Enabled = boolPtr(false)
	c.Ransomware.Enabled = boolPtr(false)
	if n := LogEffective(c); n != 2 {
		t.Fatalf("expected 2 diffs (blocklist.enabled, ransomware.enabled), got %d", n)
	}

	// Reset back to default — no diffs again.
	c = Default()
	if n := LogEffective(c); n != 0 {
		t.Fatalf("reset config should log 0 diffs, got %d", n)
	}
}
