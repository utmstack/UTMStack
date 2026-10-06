//go:build linux

package procwatch

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// TestDispatchShellActivity pins the Linux dual-emit rule: an interpreter
// execve produces process_create AND shell_activity; a non-interpreter
// produces process_create only. This test lives here (linux-only) because
// IsInterpreter is a no-op on other platforms, so the shell_activity branch
// cannot fire off-Linux and the same test would fail cross-platform.
func TestDispatchShellActivity(t *testing.T) {
	spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	tab := proctable.New()
	d := NewDispatch(tab, nil, sp, nil, nil)

	d.OnProcStart(ProcStart{PID: 100, PPID: 4, Image: "/bin/bash", Cmdline: "bash -c whoami"})
	d.OnProcStart(ProcStart{PID: 101, PPID: 4, Image: "/usr/bin/someapp", Cmdline: "someapp --run"})

	lines := readSpoolLines(t, spoolPath)
	var creates, shells int
	for _, l := range lines {
		if strings.Contains(l, `"action":"process_create"`) {
			creates++
		}
		if strings.Contains(l, `"action":"shell_activity"`) {
			shells++
		}
	}
	if creates != 2 {
		t.Fatalf("process_create events = %d, want 2 (one per execve)", creates)
	}
	if shells != 1 {
		t.Fatalf("shell_activity events = %d, want 1 (bash only, not someapp)", shells)
	}
}
