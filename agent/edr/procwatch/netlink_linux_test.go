//go:build linux

package procwatch

import (
	"encoding/binary"
	"testing"
)

// TestExecTgid exercises the pure offset extraction on synthetic buffers
// (no live kernel, no root).
func TestExecTgid(t *testing.T) {
	// A 76-byte PROC_EVENT_EXEC with tgid 1234.
	buf := make([]byte, procEventMinLen)
	binary.LittleEndian.PutUint32(buf[0:], procEventMinLen) // nlmsg_len
	binary.LittleEndian.PutUint32(buf[36:], procEventExec)  // what
	binary.LittleEndian.PutUint32(buf[52:], 4321)           // exec.process_pid
	binary.LittleEndian.PutUint32(buf[56:], 1234)           // exec.process_tgid
	if tgid, isExec := execTgid(buf, 0); !isExec || tgid != 1234 {
		t.Fatalf("execTgid = (%d, %v), want (1234, true)", tgid, isExec)
	}

	// A non-exec event (what=1) must be skipped.
	buf[36+2] = 0 // clear what
	binary.LittleEndian.PutUint32(buf[36:], 1)
	if tgid, isExec := execTgid(buf, 0); isExec || tgid != 0 {
		t.Fatalf("execTgid(non-exec) = (%d, %v), want (0, false)", tgid, isExec)
	}

	// A short buffer must be rejected.
	if tgid, isExec := execTgid(buf[:75], 0); isExec || tgid != 0 {
		t.Fatalf("execTgid(short) = (%d, %v), want (0, false)", tgid, isExec)
	}
}

// TestProcEnrichmentGoneProcess covers the /proc race: the exec'd process
// may have exited before we read it. Every helper must return the empty
// value without panicking.
func TestProcEnrichmentGoneProcess(t *testing.T) {
	// A pid that (almost certainly) does not exist on this host.
	gone := 999999999
	if ppid := procPPID(gone); ppid != 0 {
		t.Fatalf("procPPID(%d) = %d, want 0", gone, ppid)
	}
	if img := procImage(gone); img != "" {
		t.Fatalf("procImage(%d) = %q, want \"\"", gone, img)
	}
	if cl := procCmdline(gone); cl != "" {
		t.Fatalf("procCmdline(%d) = %q, want \"\"", gone, cl)
	}
}
