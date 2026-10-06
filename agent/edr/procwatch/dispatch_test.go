package procwatch

import (
	"bufio"
	"os"
	"path/filepath"
		"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

func TestDispatchAddsToTable(t *testing.T) {
	tab := proctable.New()
	d := NewDispatch(tab, nil, nil, nil, nil) // nil guard/spool/onProc => dispatch only updates the table
	d.OnProcStart(ProcStart{PID: 909, PPID: 4, Image: `C:\a.exe`})

	// table update is synchronous
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, ok := tab.Get(909); ok {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("process 909 was not added to the table")
}

func TestDispatch_InvokesOnProcAndStampsStartTS(t *testing.T) {
	tab := proctable.New()
	var got ProcStart
	d := NewDispatch(tab, nil, nil, nil, func(ps ProcStart) { got = ps })

	d.OnProcStart(ProcStart{PID: 77, PPID: 4, Image: `C:\enc.exe`, Cmdline: "enc.exe"})

	if got.PID != 77 {
		t.Fatalf("onProc not invoked with the ProcStart: %+v", got)
	}
	p, ok := tab.Get(77)
	if !ok || p.StartTS == 0 {
		t.Fatalf("StartTS not stamped in table: %+v ok=%v", p, ok)
	}
}

// readSpoolLines reads the spool file contents as a slice of lines.
func readSpoolLines(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open spool: %v", err)
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

// TestDispatchShellActivityLinux pins the dual-emit rule: an interpreter
// execve produces process_create AND shell_activity; a non-interpreter
// produces process_create only. (Linux-only: IsInterpreter is a no-op
// elsewhere, so the shell_activity branch cannot fire off-Linux.)
// TestDispatchSkipsSelfFromTelemetry pins the branding rule at the dispatch
// boundary: a procSkip-matched image must produce NO behavioral events at
// all (neither process_create nor shell_activity).
func TestDispatchSkipsSelfFromTelemetry(t *testing.T) {
	spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	tab := proctable.New()
	skip := func(image string) bool { return image == "/usr/bin/clamdscan" }
	d := NewDispatch(tab, nil, sp, skip, nil)

	d.OnProcStart(ProcStart{PID: 200, PPID: 1, Image: "/usr/bin/clamdscan", Cmdline: "clamdscan /tmp"})

	lines := readSpoolLines(t, spoolPath)
	if len(lines) != 0 {
		t.Fatalf("skipped image must emit no behavioral events, got: %v", lines)
	}
}
