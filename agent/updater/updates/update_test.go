//go:build linux

package updates

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/updater/config"
)

// The swap logic in runUpdate is tested for real, for both update targets
// (agent and EDR), against a fake systemd service per target: the "binary" is
// a shell script installed under the target's exact production name, and the
// fake service is a one-shot systemd unit (Type=idle, stays active while the
// script runs). A corrupt binary is a script that exits immediately, which
// the unit marks failed — the production failure mode the health check must
// catch and roll back.

type fakeTarget struct {
	name     string // binary name the fake target installs under
	service  string // systemd unit the fake target uses
	oldBin   string
	newBin   string
	backupBin string
}

func agentFakeTarget(base string) fakeTarget {
	oldBin := filepath.Join(base, config.ServiceFile(""))
	return fakeTarget{
		name:      "agent",
		service:   "h7updatertestagent",
		oldBin:    oldBin,
		newBin:    filepath.Join(base, config.ServiceFile("_new")),
		backupBin: filepath.Join(base, config.ServiceFile(".old")),
	}
}

func edrFakeTarget(base string) fakeTarget {
	bin := func(suffix string) string {
		return fmt.Sprintf("utmstack_edr_%s_%s%s", runtime.GOOS, runtime.GOARCH, suffix)
	}
	return fakeTarget{
		name:      "edr",
		service:   "h7updatetestedr",
		oldBin:    filepath.Join(base, bin("")),
		newBin:    filepath.Join(base, bin("_new")),
		backupBin: filepath.Join(base, bin(".old")),
	}
}

func (f fakeTarget) runUpdateTarget() updateTarget {
	return updateTarget{
		ServiceName: f.service,
		OldBin:      filepath.Base(f.oldBin),
		NewBin:      filepath.Base(f.newBin),
		BackupBin:   filepath.Base(f.backupBin),
	}
}

func systemctlCombined(args ...string) (string, error) {
	out, err := exec.Command("systemctl", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func requireSystemd(t *testing.T) {
	t.Helper()
	if _, err := os.Stat("/run/systemd/system"); err != nil {
		t.Skipf("systemd not running, skipping real service swap test: %v", err)
	}
}

// installFakeService installs and starts a one-shot unit whose ExecStart is
// the given binary, mirroring the production precondition that the service
// is running before the updater swaps its binary.
func installFakeService(t *testing.T, serviceName, binPath string) {
	t.Helper()
	unit := fmt.Sprintf(`[Unit]
Description=Updater swap test fake service (%s)
After=network.target

[Service]
Type=idle
ExecStart=%s
Restart=on-failure
RestartSec=1

[Install]
WantedBy=multi-user.target
`, serviceName, binPath)
	unitPath := "/etc/systemd/system/" + serviceName + ".service"
	if err := os.WriteFile(unitPath, []byte(unit), 0644); err != nil {
		t.Skipf("requires root to manage systemd units, skipping swap test: %v", err)
	}
	t.Cleanup(func() {
		exec.Command("systemctl", "disable", serviceName).Run()
		exec.Command("systemctl", "stop", serviceName).Run()
		os.Remove(unitPath)
		exec.Command("systemctl", "daemon-reload").Run()
	})
	if out, err := systemctlCombined("daemon-reload"); err != nil {
		t.Fatalf("cannot manage systemd units (daemon-reload: %v %s)", err, out)
	}
	if out, err := systemctlCombined("start", serviceName); err != nil {
		t.Fatalf("cannot start fake service %s: %v %s", serviceName, err, out)
	}
	time.Sleep(500 * time.Millisecond) // let the unit reach active
}

func writeWorkingBinary(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nsleep 60\n"), 0755); err != nil {
		t.Fatalf("write working binary: %v", err)
	}
}

func writeCorruptBinary(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 1\n"), 0755); err != nil {
		t.Fatalf("write corrupt binary: %v", err)
	}
}

// TestRunUpdateSwapsBinaryAndPromotesVersion proves the happy path for both
// targets: the new binary replaces the old one, the version file is
// promoted, and state files next to the binary are untouched.
func TestRunUpdateSwapsBinaryAndPromotesVersion(t *testing.T) {
	requireSystemd(t)

	baseAgent := t.TempDir()
	baseEDR := t.TempDir()
	targets := []struct {
		name string
		f    fakeTarget
	}{
		{"agent", agentFakeTarget(baseAgent)},
		{"edr", edrFakeTarget(baseEDR)},
	}

	for _, tc := range targets {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f
			base := filepath.Dir(f.oldBin)

			writeWorkingBinary(t, f.oldBin)
			// Distinguishable binaries so we can assert which one won.
			os.WriteFile(f.oldBin, []byte("#!/bin/sh\nsleep 60\n# v1\n"), 0755)
			os.WriteFile(f.newBin, []byte("#!/bin/sh\nsleep 60\n# v2\n"), 0755)
			os.WriteFile(filepath.Join(base, "version.json"), []byte(`{"version":"1.0.0","edr_version":"1.0.0"}`), 0644)
			os.WriteFile(filepath.Join(base, "version_new.json"), []byte(`{"version":"2.0.0","edr_version":"2.0.0"}`), 0644)
			// A state file (quarantine/spool/config stand-in) the swap must NOT touch.
			state := filepath.Join(base, "state.ndjson")
			os.WriteFile(state, []byte(`{"kind":"unsent"}`+"\n"), 0644)

			installFakeService(t, f.service, f.oldBin)

			oldWait := updateHealthWait
			updateHealthWait = 2 * time.Second
			defer func() { updateHealthWait = oldWait }()

			if err := runUpdate(base, f.runUpdateTarget(), true); err != nil {
				t.Fatalf("runUpdate: %v", err)
			}

			got, err := os.ReadFile(f.oldBin)
			if err != nil {
				t.Fatalf("read swapped binary: %v", err)
			}
			if !strings.Contains(string(got), "# v2") {
				t.Errorf("swapped binary is not the new one: %q", got)
			}
			v, _ := os.ReadFile(filepath.Join(base, "version.json"))
			if !strings.Contains(string(v), `"version":"2.0.0"`) {
				t.Errorf("version file not promoted: %q", v)
			}
			s, _ := os.ReadFile(state)
			if string(s) != `{"kind":"unsent"}`+"\n" {
				t.Errorf("state file was modified: %q", s)
			}
			if _, err := os.Stat(filepath.Join(base, "version.json.old")); !os.IsNotExist(err) {
				t.Errorf("version backup should be cleaned up after a healthy swap")
			}
		})
	}
}

// TestRunUpdateRollsBackOnHealthCheckFailure proves the failure path for
// both targets: a corrupt new binary makes the service die, the health check
// fails, and the old working binary + version are restored — the endpoint is
// back on the known-good version.
func TestRunUpdateRollsBackOnHealthCheckFailure(t *testing.T) {
	requireSystemd(t)

	baseAgent := t.TempDir()
	baseEDR := t.TempDir()
	targets := []struct {
		name string
		f    fakeTarget
	}{
		{"agent", agentFakeTarget(baseAgent)},
		{"edr", edrFakeTarget(baseEDR)},
	}

	for _, tc := range targets {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.f
			base := filepath.Dir(f.oldBin)

			writeWorkingBinary(t, f.oldBin)
			writeCorruptBinary(t, f.newBin)
			os.WriteFile(filepath.Join(base, "version.json"), []byte(`{"version":"1.0.0","edr_version":"1.0.0"}`), 0644)
			os.WriteFile(filepath.Join(base, "version_new.json"), []byte(`{"version":"2.0.0","edr_version":"2.0.0"}`), 0644)

			installFakeService(t, f.service, f.oldBin)

			oldWait := updateHealthWait
			updateHealthWait = 2 * time.Second
			defer func() { updateHealthWait = oldWait }()

			err := runUpdate(base, f.runUpdateTarget(), true)
			if err == nil {
				t.Fatal("runUpdate should have failed and rolled back")
			}
			if !strings.Contains(err.Error(), "rollback") {
				t.Errorf("expected rollback error, got: %v", err)
			}

			got, _ := os.ReadFile(f.oldBin)
			if !strings.Contains(string(got), "sleep 60") {
				t.Errorf("old binary was not restored: %q", got)
			}
			v, _ := os.ReadFile(filepath.Join(base, "version.json"))
			if !strings.Contains(string(v), `"version":"1.0.0"`) {
				t.Errorf("version file was not restored: %q", v)
			}
		})
	}
}

// TestEDRInstalledPredicate proves the install check: an empty dir reports
// not-installed (no download triggered), a dir with the EDR binary reports
// installed.
func TestEDRInstalledPredicate(t *testing.T) {
	base := t.TempDir()
	if edrInstalled(base) {
		t.Fatal("empty dir must not report EDR installed")
	}
	edrName := fmt.Sprintf("utmstack_edr_%s_%s", runtime.GOOS, runtime.GOARCH)
	os.WriteFile(filepath.Join(base, edrName), []byte("x"), 0755)
	if !edrInstalled(base) {
		t.Fatal("dir with EDR binary must report installed")
	}
}
