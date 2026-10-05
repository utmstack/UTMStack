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

// TestBuildSubscribeMsg pins the CN_PROC subscription byte layout against
// the kernel's struct cn_msg (cb_id{u16 idx; u16 val}, u32 seq, u32 ack,
// u16 len, u16 flags, 4 pad, u32 mcast_op). A regression to u32 idx/val
// makes the kernel see {1,0} instead of {1,1} and the subscription
// silently delivers zero events — the exact Y2.5 bug this locks out.
func TestBuildSubscribeMsg(t *testing.T) {
	buf := buildSubscribeMsg()
	if len(buf) != 40 {
		t.Fatalf("len = %d, want 40", len(buf))
	}
	checks := []struct {
		name string
		got  uint32
		want uint32
	}{
		{"nlmsg_len", binary.LittleEndian.Uint32(buf[0:]), 40},
		{"id.idx u16", uint32(binary.LittleEndian.Uint16(buf[16:])), 1},
		{"id.val u16", uint32(binary.LittleEndian.Uint16(buf[18:])), 1},
		{"seq u32", binary.LittleEndian.Uint32(buf[20:]), 1},
		{"ack u32", binary.LittleEndian.Uint32(buf[24:]), 0},
		{"cn_msg.len u16", uint32(binary.LittleEndian.Uint16(buf[28:])), 4},
		{"cn_msg.flags u16", uint32(binary.LittleEndian.Uint16(buf[30:])), 0},
		{"pad u32", binary.LittleEndian.Uint32(buf[32:]), 0},
		{"mcast_op u32", binary.LittleEndian.Uint32(buf[36:]), procCnMcastListen},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
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
