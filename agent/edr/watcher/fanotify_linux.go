//go:build linux

package watcher

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/unix"
)

// Sink is the contract the watcher hands file events to. On Linux the
// watcher feeds the orchestrator (see usn_windows.go for the Windows
// implementation of the same contract).
type Sink interface{ Enqueue(path, op string) }

// CursorStore is a placeholder for the USN cursor store; the Linux
// fanotify watcher does not persist a cursor (it is event-driven, not
// journal-driven). Kept so New's signature matches the Windows no-op.
type CursorStore interface {
	LoadUSN(volume string) (uint64, int64)
	SaveUSN(volume string, journalID uint64, nextUSN int64)
}

// PermDecider is the optional synchronous verdict source used in permission
// mode. The watcher calls Scan(path) and blocks the calling process until
// Scan returns or the deadline fires. The service wires the scanner in.
type PermDecider interface {
	Scan(path string) (malicious bool, err error)
}

// Watcher is the Linux fanotify watcher. It marks each configured volume
// (mount-level, FAN_MARK_MOUNT) and reads events from the fanotify fd,
// translating them into Sink.Enqueue calls. In permission mode it also
// answers FAN_OPEN_PERM / FAN_OPEN_EXEC_PERM events synchronously (fail-open
// on timeout/error) so the kernel does not hang the calling process.
type Watcher struct {
	volumes  []string
	sink     Sink
	excluded func(string) bool
	mode     string
	decider  PermDecider
}

// New creates a fanotify Watcher. The exported signature matches the Windows
// implementation (volumes, sink, cursors, excluded) so the service code is
// platform-agnostic. cursors is unused on Linux (fanotify is event-driven,
// not journal-driven) and is accepted only to keep the signature stable.
//
// mode is set via SetMode before Run; the default is "notify".
func New(volumes []string, sink Sink, cursors CursorStore, excluded func(string) bool) *Watcher {
	if excluded == nil {
		excluded = func(string) bool { return false }
	}
	if sink == nil {
		sink = noopSink{}
	}
	return &Watcher{volumes: volumes, sink: sink, excluded: excluded, mode: "notify"}
}

// noopSink drops events; used only if a nil sink is passed to New.
type noopSink struct{}

func (noopSink) Enqueue(string, string) {}

// SetMode selects "notify" (default) or "permission". Permission mode is only
// honored when the service runs as root AND a PermDecider is set; otherwise
// Run falls back to notify with a log line.
func (w *Watcher) SetMode(mode string) {
	switch mode {
	case "permission", "notify", "":
		if mode == "" {
			mode = "notify"
		}
		w.mode = mode
	default:
		logger.Info("UTMStack EDR watcher: unknown file_watcher_mode %q, using notify", mode)
		w.mode = "notify"
	}
}

// SetPermDecider registers the synchronous scanner used in permission mode.
// If nil when Run enters permission mode, the watcher degrades to notify.
func (w *Watcher) SetPermDecider(d PermDecider) { w.decider = d }

const (
	// permScanDeadline is the hard bound on a permission-mode scan. The kernel
	// waits for a response to a FAN_*_PERM event; if the scanner does not
	// answer in time we fail open (ALLOW) and let the async pipeline catch up.
	permScanDeadline = 1500 * time.Millisecond

	// fanReadBufSize: a single read from the fanotify fd returns one or more
	// records; 64 KiB matches the USN implementation and is enough for
	// a batch of create/modify/move events.
	fanReadBufSize = 64 * 1024
)

// fanMask is the set of events we mark on every volume. Probed against the
// target kernel (7.0, FAN_MARK_MOUNT): FAN_MODIFY, FAN_CLOSE_WRITE and
// FAN_OPEN are accepted; FAN_CREATE and FAN_MOVED_TO are rejected with EINVAL
// when marked at mount level on this kernel. A file write generates both
// FAN_MODIFY and FAN_CLOSE_WRITE, so CLOSE_WRITE alone covers "a file was
// written" with a single, clean event.
const (
	fanBaseMask = uint64(unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE)
	fanPermMask = fanBaseMask | uint64(unix.FAN_OPEN_EXEC_PERM)
	fanPermBits = uint64(unix.FAN_OPEN_EXEC_PERM)
)

// fanInitFlags / fanReportFlags are the init-time flags for notify and
// permission mode.
//
// notify: FAN_CLOEXEC + FAN_REPORT_DFID_NAME (path name).
//
// permission: same report flags plus FAN_CLASS_CONTENT so the marked
// FAN_OPEN_EXEC_PERM is delivered as a permission event we must answer.
//
// Report flags are DFID_NAME only: the kernel rejects FAN_REPORT_MNT and
// FAN_REPORT_PIDFD (fanotify_init returns EINVAL on the probe kernel), so
// they must not be requested. Volume resolution is therefore by the single
// watched root instead of by mnt_id (see processEvent).
//
// NOTE: the task's "verified" list says FAN_CLASS_CONTENT is 0x8. The actual
// value in x/sys v0.48.0 is 0x4 (0x8 is FAN_CLASS_PRE_CONTENT). We use the
// named constant so the value is never wrong.
var (
	fanInitFlagsNotify = uint(unix.FAN_CLOEXEC)
	fanInitFlagsPerm   = uint(unix.FAN_CLOEXEC | unix.FAN_CLASS_CONTENT)

	fanReportFlags = uint(unix.FAN_REPORT_DFID_NAME)
)

// volumeMark records one marked volume: its root path and the fanotify fd
// opened against it. We keep the dirFd open for the lifetime of the watch so
// the FAN_MARK_MOUNT mark stays alive (closing it would drop the mark).
type volumeMark struct {
	path  string
	dirFd int
}

// Run starts the fanotify watcher and blocks until ctx is cancelled.
//
// Setup:
//  1. Determine mode: permission only if configured AND root AND a decider
//     is set; otherwise notify.
//  2. fanotify_init with the appropriate flags.
//  3. For each volume: statfs → if a network/virtual FS, skip (the
//     "do not scan network file systems by default" rule). Otherwise open
//     the volume directory and FAN_MARK_ADD|FAN_MARK_MOUNT. Then mark the
//     EDR's own dirs (InstallDir, EngineDir, SpoolDir, QuarantineDir) with
//     FAN_MARK_IGNORE so their churn never reaches the queue (belt-and-
//     braces on top of the excluded() filter — the kernel-level ignore means
//     no event even fires, which is what prevents the processor-pegging
//     self-scan loop from the original issue).
//  4. Read loop: unix.Read on the fanotify fd, walk records, resolve the
//     path, apply excluded(), and either Enqueue (notify) or decide +
//     respond (permission).
func (w *Watcher) Run(ctx context.Context) {
	if len(w.volumes) == 0 {
		logger.Info("UTMStack EDR watcher: no volumes to watch, idle")
		<-ctx.Done()
		return
	}

	mode := w.mode
	if mode == "permission" {
		if os.Geteuid() != 0 {
			logger.Error("UTMStack EDR watcher: permission mode requires root (uid 0); falling back to notify")
			mode = "notify"
		} else if w.decider == nil {
			logger.Error("UTMStack EDR watcher: permission mode has no PermDecider; falling back to notify")
			mode = "notify"
		}
	}
	logger.Info("UTMStack EDR watcher: starting in %s mode over %d volume(s)", mode, len(w.volumes))

	var initFlags uint
	var mask uint64
	if mode == "permission" {
		initFlags = fanInitFlagsPerm
		mask = fanPermMask
	} else {
		initFlags = fanInitFlagsNotify
		mask = fanBaseMask
	}

	fd, err := unix.FanotifyInit(initFlags, fanReportFlags)
	if err != nil {
		logger.Error("UTMStack EDR watcher: fanotify_init: %v", err)
		return
	}
	defer unix.Close(fd)

	// Mark volumes.
	var marks []volumeMark
	defer func() {
		for _, m := range marks {
			unix.Close(m.dirFd)
		}
	}()

	for _, vol := range w.volumes {
		supported, reason := isSupportedVolume(vol)
		if !supported {
			logger.Info("UTMStack EDR watcher: skipping %s (%s)", vol, reason)
			continue
		}
		dirFd, err := unix.Open(vol, unix.O_RDONLY|unix.O_DIRECTORY, 0)
		if err != nil {
			logger.Error("UTMStack EDR watcher: open %s: %v", vol, err)
			continue
		}
		if err := unix.FanotifyMark(fd, unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT, mask, dirFd, ""); err != nil {
			logger.Error("UTMStack EDR watcher: mark %s: %v", vol, err)
			unix.Close(dirFd)
			continue
		}
		marks = append(marks, volumeMark{path: vol, dirFd: dirFd})
	}

	if len(marks) == 0 {
		logger.Error("UTMStack EDR watcher: no volume could be marked, idle")
		<-ctx.Done()
		return
	}
	logger.Info("UTMStack EDR watcher: marked %d volume(s)", len(marks))

	// Belt-and-braces: mark the EDR's own dirs as ignored at the kernel level
	// so their churn never reaches the queue. This is the hard rule from the
	// original issue — scanning our own quarantine store / engine dir caused a
	// processor-pegging loop and a branding leak. The excluded() filter is a
	// second layer on top (see processEvents).
	for _, m := range marks {
		for _, own := range ownDirs() {
			if own == "" {
				continue
			}
			// Only ignore if own is under the marked volume.
			if !isUnder(own, m.path) {
				continue
			}
			ownFd, err := unix.Open(own, unix.O_RDONLY|unix.O_DIRECTORY, 0)
			if err != nil {
				// own may not exist yet; that's fine, we just can't ignore it
				// at the kernel level (the excluded() filter still covers it).
				continue
			}
			// FAN_MARK_IGNORE on a non-inode mark (a directory fd) requires
			// FAN_MARK_IGNORED_SURV_MODIFY; without it fanotify_mark returns
			// EINVAL and the ignore silently fails. FAN_MARK_IGNORE_SURV is the
			// IGNORE|IGNORED_SURV_MODIFY convenience pair.
			if err := unix.FanotifyMark(fd, unix.FAN_MARK_IGNORE_SURV|unix.FAN_MARK_MOUNT, mask, ownFd, ""); err != nil {
				logger.Error("UTMStack EDR watcher: ignore-mark %s: %v", own, err)
			}
			unix.Close(ownFd)
		}
	}

	// Read loop: poll + non-blocking read so ctx cancellation is observed
	// promptly (a plain blocking read would hang the goroutine until an event).
	if err := unix.SetNonblock(fd, true); err != nil {
		logger.Error("UTMStack EDR watcher: setnonblock: %v", err)
		return
	}
	var pfd [1]unix.PollFd
	pfd[0].Fd = int32(fd)
	pfd[0].Events = unix.POLLIN
	buf := make([]byte, fanReadBufSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		n, err := unix.Poll(pfd[:], 500)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			logger.Error("UTMStack EDR watcher: poll: %v", err)
			return
		}
		if n == 0 {
			continue
		}
		got, rerr := unix.Read(fd, buf)
		if rerr != nil {
			if errors.Is(rerr, syscall.EAGAIN) || errors.Is(rerr, syscall.EWOULDBLOCK) {
				continue
			}
			logger.Error("UTMStack EDR watcher: read: %v", rerr)
			return
		}
		if got <= 0 {
			continue
		}
		events := parseFanotifyEvents(buf[:got])
		for _, ev := range events {
			w.processEvent(fd, ev)
		}
	}
}

// isUnder reports whether path is under root (path == root or path starts
// with root + sep).
func isUnder(path, root string) bool {
	if path == root {
		return true
	}
	return len(path) > len(root) && strings.HasPrefix(path, root) && path[len(root)] == os.PathSeparator
}

// ownDirs returns the EDR's own directories that must never be scanned.
// These are the install tree, engine, spool, and quarantine dirs.
func ownDirs() []string {
	return []string{config.InstallDir, config.EngineDir, config.SpoolDir, config.QuarantineDir}
}

// isSupportedVolume classifies the filesystem at path by its superblock magic
// and reports whether it is safe to mark. Network and virtual filesystems
// (NFS, CIFS, 9P, overlay) are skipped by default — this is the "do not scan
// network file systems by default" rule from the issue.
//
// Returns (true, "") for local filesystems, (false, reason) for the
// ones we refuse to mark. The reason is for logging only.
func isSupportedVolume(path string) (bool, string) {
	var sf unix.Statfs_t
	if err := unix.Statfs(path, &sf); err != nil {
		return false, fmt.Sprintf("statfs: %v", err)
	}
	return classifyMagic(sf.Type), magicName(sf.Type)
}

// classifyMagic is the pure magic-number classification (testable without a
// live filesystem). Network and virtual filesystems must not be marked.
// NOTE: NFS and 9P2000 (plan 9) share the magic number 0x6969, so a single
// case covers both.
func classifyMagic(magic int64) bool {
	switch magic {
	case unix.NFS_SUPER_MAGIC: // NFS / 9P2000 (both 0x6969)
		return false
	case unix.CIFS_SUPER_MAGIC: // CIFS/SMB
		return false
	case unix.OVERLAYFS_SUPER_MAGIC:
		return false
	default:
		// ext4, xfs, btrfs, tmpfs, vfat, etc. — local, mark it.
		return true
	}
}

// magicName is a short human label for logging; best-effort.
func magicName(magic int64) string {
	switch magic {
	case unix.NFS_SUPER_MAGIC:
		return "nfs/9p"
	case unix.CIFS_SUPER_MAGIC:
		return "cifs"
	case unix.OVERLAYFS_SUPER_MAGIC:
		return "overlayfs"
	default:
		return fmt.Sprintf("magic 0x%x", magic)
	}
}

// fanotifyEvent is one decoded event from the read buffer.
type fanotifyEvent struct {
	mask    uint64
	fd      int32
	mntID   uint64
	relName string // name from DFID_NAME, relative to the mark root
	hasName bool
	isPerm  bool
}

// parseFanotifyEvents walks a read buffer containing one or more fanotify
// records. Each record starts with struct fanotify_event_metadata (24 bytes)
// followed by metadata_len bytes of stacked info records. We extract the
// mask, fd, mnt_id, and the DFID_NAME (file handle + name).
//
// The name offset is NOT fixed: it comes after the file_handle inside the
// DFID_NAME info record. We read handle_bytes and skip the handle to find it.
func parseFanotifyEvents(buf []byte) []fanotifyEvent {
	var out []fanotifyEvent
	off := 0
	for off < len(buf) {
		if off+24 > len(buf) {
			break
		}
		meta := fanotifyMeta{
			eventLen: binary.LittleEndian.Uint32(buf[off:]),
			mask:     binary.LittleEndian.Uint64(buf[off+8:]),
			fd:       int32(binary.LittleEndian.Uint32(buf[off+16:])),
		}
		if meta.eventLen < 24 || int(off)+int(meta.eventLen) > len(buf) {
			break
		}
		ev := fanotifyEvent{mask: meta.mask, fd: meta.fd, isPerm: meta.mask&fanPermBits != 0}
		infoStart := off + 24
		infoEnd := off + int(meta.eventLen)
		for i := infoStart; i+4 <= infoEnd; {
			// struct fanotify_event_info_header: {u8 type; u8 pad; u16 len}
			iType := buf[i]
			iLen := int(binary.LittleEndian.Uint16(buf[i+2:]))
			if iLen < 4 || i+iLen > infoEnd {
				break
			}
			switch iType {
			case unix.FAN_EVENT_INFO_TYPE_DFID_NAME:
				// struct fanotify_event_info_fid:
				//   header (4) + fsid (8) + file_handle { u32 handle_bytes;
				//   i32 handle_type; u8 f_handle[handle_bytes] } + name (null-terminated)
				if i+12 > infoEnd {
					break
				}
				handleBytes := int(binary.LittleEndian.Uint32(buf[i+12:]))
				nameStart := i + 20 + handleBytes
				if nameStart > infoEnd {
					break
				}
				nameEnd := nameStart
				for nameEnd < infoEnd && buf[nameEnd] != 0 {
					nameEnd++
				}
				ev.relName = string(buf[nameStart:nameEnd])
				ev.hasName = true
			case unix.FAN_EVENT_INFO_TYPE_MNT:
				if i+8 > infoEnd {
					break
				}
				ev.mntID = binary.LittleEndian.Uint64(buf[i+4:])
			case unix.FAN_EVENT_INFO_TYPE_PIDFD:
				// We don't need the pidfd for the watcher's decisions; skip.
			}
			// Info records are 8-byte aligned.
			i += (iLen + 7) &^ 7
		}
		out = append(out, ev)
		off += (int(meta.eventLen) + 7) &^ 7
	}
	return out
}

// fanotifyMeta mirrors struct fanotify_event_metadata.
type fanotifyMeta struct {
	eventLen uint32
	mask     uint64
	fd       int32
}

// processEvent resolves the event's path, applies the excluded() filter, and
// either enqueues it (notify) or decides + responds (permission).
//
// Path resolution: the DFID_NAME info carries the file name relative to the
// marked volume root; we join the watched root with it. Submounts under the
// The watcher watches a single root (/ on Linux), so the volume is
// w.volumes[0]; len(w.volumes) > 0 is guaranteed by Run's early return.
func (w *Watcher) processEvent(fd int, ev fanotifyEvent) {
	// Overflow is not a file event; log and drop.
	if ev.mask&unix.FAN_Q_OVERFLOW != 0 {
		logger.Error("UTMStack EDR watcher: fanotify queue overflow, events dropped")
		return
	}
	if len(w.volumes) == 0 {
		return
	}
	vol := w.volumes[0]

	// Resolve the path. With FAN_MARK_MOUNT the kernel does not deliver the
	// DFID_NAME (metadata_len is just the 24-byte header), so the event's
	// path comes from readlink-ing the file descriptor it carries — the
	// documented resolution method for mount-level marks. Events without a
	// usable fd (FAN_NOFD) fall back to the relative name when present.
	path, ok := w.eventPath(vol, ev)
	if !ok {
		return
	}

	// Permission events MUST be answered before any other branch may return:
	// the calling process sleeps in the kernel (fanotify_get_response) until
	// we write the response, so an early return here hangs it forever. In
	// particular the EDR's own management CLI lives inside InstallDir (an
	// excluded path) — excluding it must mean "allow without scanning", not
	// "no response".
	if ev.isPerm {
		allow := true
		if w.decider != nil && !(w.excluded != nil && w.excluded(path)) {
			allow = w.permDecision(path)
		}
		// ALWAYS respond — an unanswered permission event hangs the calling
		// process in the kernel. The response is a write(2) of
		// struct fanotify_response to the fanotify fd; there is no
		// FANOTIFY_RESPONSE syscall in x/sys v0.48.0.
		resp := uint32(unix.FAN_ALLOW)
		if !allow {
			resp = uint32(unix.FAN_DENY)
		}
		if ev.fd != unix.FAN_NOFD {
			r := unix.FanotifyResponse{Fd: ev.fd, Response: resp}
			if _, err := unix.Write(fd, unsafe.Slice((*byte)(unsafe.Pointer(&r)), unsafe.Sizeof(r))); err != nil {
				logger.Error("UTMStack EDR watcher: fanotify response: %v", err)
			}
			// Close the event fd so the kernel does not hold it open (the man
			// page says the fd must be closed; for permission events the
			// response implies it, but closing is defensive and correct).
			unix.Close(int(ev.fd))
		}
		if !allow {
			// Malicious verdict: enqueue so the normal detection/quarantine
			// event flow fires. The scanner already quarantines on its own
			// verdict path (ScanFile → Quarantine); enqueuing here makes the
			// event appear. The orchestrator's inflight coalescing
			// (orchestrator.go: Enqueue) dedupes if the scan also enqueued.
			w.sink.Enqueue(path, "create")
		}
		return
	}

	// Application-level exclusion (second layer; the kernel-level FAN_MARK_IGNORE
	// already covers the EDR's own dirs).
	if w.excluded != nil && w.excluded(path) {
		return
	}

	// Notify mode: map mask → op. The event fd (if any) must be closed — the
	// kernel opened it for us; leaking it exhausts fds under churn.
	op := fanOp(ev.mask)
	if ev.fd != unix.FAN_NOFD {
		unix.Close(int(ev.fd))
	}
	if op == "" {
		return
	}
	w.sink.Enqueue(path, op)
}

// eventPath resolves the absolute path for an event. With FAN_MARK_MOUNT the
// kernel delivers a file fd but no DFID_NAME (the metadata is just the 24-byte
// header), so the documented resolution is readlink-ing /proc/self/fd/<fd>.
// Events without a usable fd (FAN_NOFD) fall back to the relative name joined
// to the volume root — the hermetic tests use fd=FAN_NOFD + relName.
func (w *Watcher) eventPath(vol string, ev fanotifyEvent) (string, bool) {
	if ev.fd != unix.FAN_NOFD {
		if target, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", ev.fd)); err == nil {
			return target, true
		}
		// The fd may be stale (the file was deleted between the event and the
		// readlink); fall through to the name if we have one.
	}
	if ev.hasName {
		return filepath.Join(vol, ev.relName), true
	}
	return "", false
}

// permDecision asks the decider for a verdict with a hard deadline. Any
// timeout or error fails open (returns true = ALLOW).
func (w *Watcher) permDecision(path string) (allow bool) {
	type result struct {
		mal bool
		err error
	}
	ch := make(chan result, 1)
	go func() {
		m, e := w.decider.Scan(path)
		ch <- result{m, e}
	}()
	select {
	case <-time.After(permScanDeadline):
		logger.Error("UTMStack EDR watcher: permission scan of %s timed out (%v), failing open", path, permScanDeadline)
		return true
	case r := <-ch:
		if r.err != nil {
			logger.Error("UTMStack EDR watcher: permission scan of %s: %v, failing open", path, r.err)
			return true
		}
		return !r.mal
	}
}

// fanOp maps a fanotify event mask to the orchestrator's op string.
// CLOSE_WRITE / MODIFY mean "a file was written" — the pipeline's existing
// "create" op is the closest match (a brand-new file's first write is the
// case we care about; re-writes of known files coalesce in the inflight map).
// FAN_CREATE / FAN_MOVED_TO are not in the mark mask on this kernel (see
// fanBaseMask comment), so no case for them.
func fanOp(mask uint64) string {
	switch {
	case mask&unix.FAN_CLOSE_WRITE != 0:
		return "create"
	case mask&unix.FAN_MODIFY != 0:
		return "modify"
	default:
		return ""
	}
}
