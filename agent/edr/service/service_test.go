package service

import (
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/orchestrator"
)

func TestSigDBVersionStripsEngineName(t *testing.T) {
	// clamd VERSION banner form; must yield only the sig-db version, no engine name.
	got := sigDBVersion("ClamAV 1.0.1/27000/Wed Jan 10 09:00:00 2024")
	if got != "27000" {
		t.Fatalf("sigDBVersion = %q, want 27000", got)
	}
	if strings.Contains(strings.ToLower(got), "clam") {
		t.Fatalf("sig-db version leaked engine name: %q", got)
	}
	if v := sigDBVersion("garbage-no-slashes"); v != "" {
		t.Fatalf("expected empty for malformed banner, got %q", v)
	}
}

func TestDefaultTrustedProcessesMatchesFPWorkloads(t *testing.T) {
	tr := orchestrator.NewExcluder(defaultTrustedProcesses())
	trusted := []string{
		`C:\Program Files\Veeam\Backup\Veeam.Backup.Manager.exe`, // subtree
		`C:\Program Files (x86)\Acronis\TrueImage\service.exe`,
		`C:\Windows\System32\vssvc.exe`,   // OS VSS service (full path)
		`C:\Program Files\Microsoft OneDrive\OneDrive.exe`,
	}
	for _, p := range trusted {
		if !tr.Excluded(p) {
			t.Errorf("expected built-in trust to cover %q", p)
		}
	}
	// Must NOT trust the T1490 tells themselves, or arbitrary user binaries.
	for _, p := range []string{
		`C:\Windows\System32\vssadmin.exe`,
		`C:\Windows\System32\wbadmin.exe`,
		`C:\Users\victim\Downloads\enc.exe`,
	} {
		if tr.Excluded(p) {
			t.Errorf("built-in trust must NOT cover %q", p)
		}
	}
}
