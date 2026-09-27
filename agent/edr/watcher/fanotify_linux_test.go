//go:build linux

package watcher

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// TestIsSupportedVolumeMagic classifies superblock magic numbers via the pure
// helper. The same numeric test doubles the real unix.* constants so a future
// change to a magic value (or a mis-assignment) is caught. NFS and 9P share 0x6969.
func TestIsSupportedVolumeMagic(t *testing.T) {
	cases := []struct {
		magic int64
		want  bool // supported (should be marked)
	}{
		{int64(unix.NFS_SUPER_MAGIC), false},
		{int64(unix.CIFS_SUPER_MAGIC), false},
		{int64(unix.OVERLAYFS_SUPER_MAGIC), false},
		{int64(unix.TMPFS_MAGIC), true},
		{0xef53, true},     // ext4
		{0x9123683e, true}, // btrfs
		{0x58465342, true}, // xfs
		{0x6969, false},    // 9P2000 / NFS
	}
	for _, c := range cases {
		if got := classifyMagic(c.magic); got != c.want {
			t.Errorf("classifyMagic(0x%x) = %v, want %v", c.magic, got, c.want)
		}
	}
}

// TestIsSupportedVolumeLive probes the real filesystem at / and (if it is not
// a network FS) at t.TempDir's backing filesystem.
func TestIsSupportedVolumeLive(t *testing.T) {
	ok, reason := isSupportedVolume("/")
	if !ok {
		t.Fatalf("root / should be a supported local volume, got skip reason %q", reason)
	}
	// tmpdir may be on a different mount (tmpfs is supported).
	tmp := t.TempDir()
	ok2, reason2 := isSupportedVolume(tmp)
	if !ok2 {
		t.Fatalf("temp dir %s should be supported, got skip reason %q", tmp, reason2)
	}
}

// buildFanRecord encodes one fanotify event record (metadata + DFID_NAME info)
// shaped exactly like the kernel would deliver it.
func buildFanRecord(mask uint64, fd int32, name string) []byte {
	nameBytes := []byte(name)

	// DFID_NAME info: header(4) + fsid(8) + file_handle{u32 handle_bytes;
	// i32 handle_type; u8 f_handle[hb]} + null-terminated name.
	const handleBytes = 12
	fidBody := 8 + 4 + 4 + handleBytes + len(nameBytes) + 1 // fsid + handle hdr + name + NUL
	fidRec := make([]byte, (4+fidBody+7)&^7)
	fidRec[0] = unix.FAN_EVENT_INFO_TYPE_DFID_NAME
	binary.LittleEndian.PutUint16(fidRec[2:], uint16(4+fidBody))
	// fsid zeroed (4 u32 = 16 bytes? no: __kernel_fsid_t is two u32 = 8 bytes)
	// file_handle
	binary.LittleEndian.PutUint32(fidRec[12:], uint32(handleBytes)) // handle_bytes
	// handle_type left 0; f_handle[handleBytes] left 0
	copy(fidRec[20+handleBytes:], nameBytes)

	infos := fidRec

	meta := 24
	recLen := (meta + len(infos) + 7) &^ 7
	rec := make([]byte, recLen)
	binary.LittleEndian.PutUint32(rec[0:], uint32(recLen))
	rec[4] = 3 // vers
	binary.LittleEndian.PutUint16(rec[8:], uint16(len(infos)))
	binary.LittleEndian.PutUint64(rec[8:], mask)
	binary.LittleEndian.PutUint32(rec[16:], uint32(fd))
	copy(rec[24:], infos)
	return rec
}

// TestParseFanotifyEvents feeds synthetic records (no real kernel events) and
// asserts the decoded mask/name.
func TestParseFanotifyEvents(t *testing.T) {
	closeWrite := buildFanRecord(unix.FAN_CLOSE_WRITE, -1, "sub/a.txt")
	modify := buildFanRecord(unix.FAN_MODIFY, -1, "sub/a.txt")
	perm := buildFanRecord(unix.FAN_OPEN_EXEC_PERM, 7, "evil.exe")
	buf := append(append(append([]byte{}, closeWrite...), modify...), perm...)

	events := parseFanotifyEvents(buf)
	if len(events) != 3 {
		t.Fatalf("got %d events, want 3", len(events))
	}
	if events[0].relName != "sub/a.txt" || !events[0].hasName {
		t.Fatalf("bad close_write: %+v", events[0])
	}
	if events[0].mask&unix.FAN_CLOSE_WRITE == 0 {
		t.Fatalf("bad close_write mask: %#x", events[0].mask)
	}
	if events[1].mask&unix.FAN_MODIFY == 0 {
		t.Fatalf("bad modify mask: %#x", events[1].mask)
	}
	if !events[2].isPerm || events[2].fd != 7 || events[2].relName != "evil.exe" {
		t.Fatalf("bad perm event: %+v", events[2])
	}
}

// TestProcessEventOverflow verifies a FAN_Q_OVERFLOW record logs/drops without
// touching the sink.
func TestProcessEventOverflow(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	ev := fanotifyEvent{mask: unix.FAN_Q_OVERFLOW, fd: -1}
	w.processEvent(-1, ev)
	if len(sk.calls) != 0 {
		t.Fatalf("overflow should not enqueue, got %v", sk.calls)
	}
}

// TestParseFanotifyEventsTruncated ensures a truncated record does not panic
// and yields no usable event.
func TestParseFanotifyEventsTruncated(t *testing.T) {
	good := buildFanRecord(unix.FAN_CLOSE_WRITE, -1, "x")
	if evs := parseFanotifyEvents(good[:10]); len(evs) != 0 {
		t.Fatalf("truncated record should yield 0 events, got %d", len(evs))
	}
}

type recSink struct {
	calls [][2]string
}

func (r *recSink) Enqueue(p, op string) { r.calls = append(r.calls, [2]string{p, op}) }

// TestProcessEventExcluded verifies the application-level exclusion short-
// circuits the enqueue.
func TestProcessEventExcluded(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	excluded := func(p string) bool { return filepath.Base(p) == "skip.txt" }
	w := New([]string{dir}, sk, nil, excluded)
	ev := fanotifyEvent{mask: unix.FAN_CLOSE_WRITE, fd: -1, relName: "skip.txt", hasName: true}
	w.processEvent(-1, ev)
	if len(sk.calls) != 0 {
		t.Fatalf("excluded path should not enqueue, got %v", sk.calls)
	}
}

// TestProcessEventNotify verifies the mask→op mapping and path join.
func TestProcessEventNotify(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	ev := fanotifyEvent{mask: unix.FAN_CLOSE_WRITE, fd: -1, relName: "sub/a.txt", hasName: true}
	w.processEvent(-1, ev)
	if len(sk.calls) != 1 {
		t.Fatalf("want 1 call, got %v", sk.calls)
	}
	if sk.calls[0][0] != filepath.Join(dir, "sub/a.txt") || sk.calls[0][1] != "create" {
		t.Fatalf("bad call: %v", sk.calls[0])
	}
}

// TestProcessEventPermissionMalicious verifies a malicious decider denies and
// enqueues the path so the async pipeline fires the detection.
func TestProcessEventPermissionMalicious(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	w.SetMode("permission")
	w.SetPermDecider(&funcDecider{mal: true})
	ev := fanotifyEvent{mask: unix.FAN_OPEN_EXEC_PERM, relName: "bad.exe", hasName: true, isPerm: true, fd: -1}
	w.processEvent(-1, ev) // fd -1: response write skipped
	if len(sk.calls) != 1 || sk.calls[0][1] != "create" {
		t.Fatalf("malicious perm event should enqueue create, got %v", sk.calls)
	}
}

// TestProcessEventPermissionClean verifies a clean decider allows and does not
// enqueue.
func TestProcessEventPermissionClean(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	w.SetMode("permission")
	w.SetPermDecider(&funcDecider{mal: false})
	ev := fanotifyEvent{mask: unix.FAN_OPEN_EXEC_PERM, relName: "ok.exe", hasName: true, isPerm: true, fd: -1}
	w.processEvent(-1, ev)
	if len(sk.calls) != 0 {
		t.Fatalf("clean perm event should not enqueue, got %v", sk.calls)
	}
}

// TestProcessEventPermissionFailOpen verifies a decider error fails open
// (allow) and does not enqueue.
func TestProcessEventPermissionFailOpen(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	w.SetMode("permission")
	w.SetPermDecider(&funcDecider{err: os.ErrNotExist})
	ev := fanotifyEvent{mask: unix.FAN_OPEN_EXEC_PERM, relName: "e.exe", hasName: true, isPerm: true, fd: -1}
	w.processEvent(-1, ev)
	if len(sk.calls) != 0 {
		t.Fatalf("fail-open perm event should not enqueue, got %v", sk.calls)
	}
}

// TestProcessEventPermissionDegradesWithoutDecider verifies that permission
// mode with a nil decider degrades to notify-only (allow everything, no
// panic) rather than blocking.
func TestProcessEventPermissionDegradesWithoutDecider(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	w := New([]string{dir}, sk, nil, nil)
	w.SetMode("permission")
	// No SetPermDecider call: w.decider is nil.
	if w.decider != nil {
		t.Fatalf("expected nil decider")
	}
	ev := fanotifyEvent{mask: unix.FAN_OPEN_EXEC_PERM, relName: "d.exe", hasName: true, isPerm: true, fd: -1}
	// Must not panic; degrades to allow (no enqueue).
	w.processEvent(-1, ev)
	if len(sk.calls) != 0 {
		t.Fatalf("degraded perm event should not enqueue, got %v", sk.calls)
	}
}

// funcDecider is a test PermDecider.
type funcDecider struct {
	mal    bool
	err    error
	called int
}

func (f *funcDecider) Scan(string) (bool, error) { f.called++; return f.mal, f.err }

// TestProcessEventExcludedPermissionStillResponds is the regression test for the
// self-intercept hang: when the EDR's own CLI (an excluded path, e.g. inside
// InstallDir) is exec'd in permission mode, the watcher MUST still write the
// kernel response — excluding means "allow without scanning", not "no response".
// A missing response leaves the calling process asleep in fanotify_get_response
// forever. We use a pipe as the fanotify fd so we can read the response bytes.
func TestProcessEventExcludedPermissionStillResponds(t *testing.T) {
	dir := t.TempDir()
	sk := &recSink{}
	excluded := func(p string) bool { return strings.Contains(p, "skip") }
	dec := &funcDecider{mal: true} // would DENY if wrongly consulted
	w := New([]string{dir}, sk, nil, excluded)
	w.SetMode("permission")
	w.SetPermDecider(dec)

	r, pw, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer pw.Close()

	// fd arg = pipe write end (where the response is written). ev.fd is a dummy
	// (the code closes ev.fd after responding; 42 is an unused, unopen fd).
	ev := fanotifyEvent{mask: unix.FAN_OPEN_EXEC_PERM, fd: 42, relName: "skip.bin", hasName: true, isPerm: true}
	w.processEvent(int(pw.Fd()), ev)

	// struct fanotify_response { int32 fd; uint32 response; } = 8 bytes.
	var resp [8]byte
	if _, err := unix.Read(int(r.Fd()), resp[:]); err != nil {
		t.Fatalf("no kernel response written for excluded perm event: %v", err)
	}
	if got := binary.LittleEndian.Uint32(resp[4:]); got != unix.FAN_ALLOW {
		t.Fatalf("excluded perm event must FAN_ALLOW, got %#x", got)
	}
	if dec.called != 0 {
		t.Fatalf("excluded path must skip the decider, but it was consulted %d times", dec.called)
	}
	if len(sk.calls) != 0 {
		t.Fatalf("excluded perm event must not enqueue, got %v", sk.calls)
	}
}

// TestSetModeNormalization verifies unknown modes fall back to notify.
func TestSetModeNormalization(t *testing.T) {
	w := New(nil, nil, nil, nil)
	w.SetMode("permission")
	if w.mode != "permission" {
		t.Fatalf("mode: %q", w.mode)
	}
	w.SetMode("")
	if w.mode != "notify" {
		t.Fatalf("empty mode should default to notify, got %q", w.mode)
	}
	w.SetMode("bogus")
	if w.mode != "notify" {
		t.Fatalf("bogus mode should fall back to notify, got %q", w.mode)
	}
}

// TestIsUnder verifies the kernel-ignore scoping helper.
func TestIsUnder(t *testing.T) {
	cases := []struct {
		path, root string
		want       bool
	}{
		{"/a/b", "/a", true},
		{"/a", "/a", true},
		{"/a/b", "/b", false},
		{"/ab", "/a", false}, // prefix, not under
	}
	for _, c := range cases {
		if got := isUnder(c.path, c.root); got != c.want {
			t.Errorf("isUnder(%q, %q) = %v, want %v", c.path, c.root, got, c.want)
		}
	}
}

// TestFanotifyLive is a real kernel integration test. It requires CAP_SYS_ADMIN
// (fanotify_init on most kernels), so it skips cleanly when unavailable.
// Run with sudo to exercise it: sudo go test -run TestFanotifyLive ./edr/watcher
func TestFanotifyLive(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("fanotify requires root (CAP_SYS_ADMIN); run with sudo")
	}
	dir := t.TempDir()
	fd, err := unix.FanotifyInit(unix.FAN_CLOEXEC, fanReportFlags)
	if err != nil {
		t.Skipf("fanotify_init: %v", err)
	}
	defer unix.Close(fd)

	dfd, err := unix.Open(dir, unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		t.Fatalf("open %s: %v", dir, err)
	}
	defer unix.Close(dfd)

	if err := unix.FanotifyMark(fd, unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT, fanBaseMask, dfd, ""); err != nil {
		t.Fatalf("fanotify_mark: %v", err)
	}

	// Write a file to generate a FAN_CLOSE_WRITE.
	target := filepath.Join(dir, "live_probe.txt")
	if err := os.WriteFile(target, []byte("x"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// Read events with a 2s timeout. With FAN_MARK_MOUNT the kernel delivers
	// the file's fd but no name (metadata is just the 24-byte header), so
	// resolve each event's path via readlink — the same method production uses.
	// Keep reading until we find the event for our specific file or timeout.
	unix.SetNonblock(fd, true)
	deadline := time.Now().Add(2 * time.Second)
	var gotPath string
	buf := make([]byte, fanReadBufSize)
	for time.Now().Before(deadline) {
		n, rerr := unix.Read(fd, buf)
		if rerr != nil || n == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		for _, ev := range parseFanotifyEvents(buf[:n]) {
			if ev.fd != unix.FAN_NOFD {
				target, _ := os.Readlink("/proc/self/fd/" + fmt.Sprintf("%d", ev.fd))
				unix.Close(int(ev.fd))
				if strings.Contains(target, "live_probe.txt") {
					gotPath = target
					break
				}
			}
		}
		if gotPath != "" {
			break
		}
	}
	if gotPath == "" {
		t.Fatalf("no fanotify event received within 2s")
	}
	// The readlink target is an absolute path to the file we wrote.
	if !strings.Contains(gotPath, "live_probe.txt") {
		t.Fatalf("expected path to contain %q, got %q", "live_probe.txt", gotPath)
	}
	t.Logf("live fanotify event for %s", gotPath)
}

// TestNewDefaults verifies New fills sensible defaults (excluded, sink, mode).
func TestNewDefaults(t *testing.T) {
	w := New(nil, nil, nil, nil)
	if w.mode != "notify" {
		t.Fatalf("default mode: %q", w.mode)
	}
	if w.excluded == nil || w.sink == nil {
		t.Fatalf("defaults not set: excluded==nil: %v sink==nil: %v", w.excluded == nil, w.sink == nil)
	}
}
