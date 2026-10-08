//go:build linux

package ransomware

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/unix"
)

// desiredFeedBits are the fanotify event bits the feed wants to mark on every
// volume. They are probed at runtime, one per FAN_MARK_ADD, because
// FAN_MARK_MOUNT rejects bits the target kernel cannot honor: on the kernel
// observed here FAN_CREATE and FAN_MOVED_TO are declined with EINVAL at mount
// level, while FAN_MODIFY and FAN_CLOSE_WRITE are accepted. Fanotify OR-s the
// masks of marks over the same object, so marking each accepted bit
// individually is equivalent to marking their union, and the feed degrades to
// write-only on kernels without the extended bits rather than failing to start.
//
// FAN_DELETE is deliberately omitted: without FAN_REPORT_DFID_NAME the kernel
// gives it no fd, so its path is unresolvable and the event would be dropped in
// processFeedEvent anyway — marking it only costs a syscall. CREATE and
// MOVED_TO DO carry a fd (the new destination), which is exactly what the
// churn and ransom-note sensors need. The list is ordered so the always-
// available write bits come first: a kernel that accepts nothing still yields
// them.
var desiredFeedBits = []uint64{
	unix.FAN_CLOSE_WRITE,
	unix.FAN_MODIFY,
	unix.FAN_CREATE,
	unix.FAN_MOVED_TO,
}

// fullFeedMask is the union of every desired bit, i.e. what the feed requests
// from a fully-capable kernel. The mark path uses it to decide whether to log
// a degraded mark.
var fullFeedMask = desiredFeedBits[0] | desiredFeedBits[1] | desiredFeedBits[2] | desiredFeedBits[3]

// fanReadBufSize: a single read from the fanotify fd returns one or more
// records; 64 KiB matches the watcher and is enough for a batch of events.
const fanReadBufSize = 64 * 1024

// fanotifyFeed is the Linux fanotify file-activity feed. It marks each
// configured volume at mount level (FAN_MARK_MOUNT) with the largest mask the
// kernel accepts — writes always, plus create/moved-to when available — and
// delivers per-process FileEvents (PID from FAN_REPORT_PIDFD, path from
// readlink of the event fd) to the guard's sink. It is observational only (no
// FAN_CLASS_CONTENT): the main watcher already owns permission mode.
type fanotifyFeed struct {
	volumes []string
}

// NewFeed returns the Linux fanotify file-activity feed over the given
// volumes. With no volumes Run blocks until ctx is cancelled (idle).
func NewFeed(vols ...string) FileActivityFeed { return fanotifyFeed{volumes: vols} }

// Run starts the fanotify feed and blocks until ctx is cancelled.
//
// Setup:
//  1. fanotify_init with FAN_CLOEXEC|FAN_REPORT_PIDFD|FAN_UNLIMITED_QUEUE.
//     FAN_UNLIMITED_QUEUE stops the default 16-slot kernel queue from
//     overflowing (and silently dropping events) on busy mounts.
//     FAN_REPORT_DFID_NAME must NOT be combined with FAN_REPORT_PIDFD — the
//     kernel returns ENOSYS. Path resolution is by readlink of the event fd.
//  2. For each volume: statfs → skip network/virtual FS (same magic
//     classification as the watcher). Otherwise open the volume directory
//     and FAN_MARK_ADD|FAN_MARK_MOUNT. dirFds stay open for the watch
//     lifetime (closing drops the mark).
//  3. Read loop: unix.Poll (500ms) + non-blocking unix.Read, parse
//     fanotify_event_metadata records, resolve the path, and deliver.
//
// Returns ctx.Err() on clean shutdown; returns the error on init/mark/read
// failure so Guard.Run's retry loop restarts the feed.
func (f fanotifyFeed) Run(ctx context.Context, sink func(FileEvent)) error {
	if len(f.volumes) == 0 {
		logger.Info("UTMStack EDR: ransomware feed: no volumes to watch, idle")
		<-ctx.Done()
		return ctx.Err()
	}

	fd, err := unix.FanotifyInit(unix.FAN_CLOEXEC|unix.FAN_REPORT_PIDFD|unix.FAN_UNLIMITED_QUEUE, 0)
	if err != nil {
		return fmt.Errorf("fanotify_init: %w", err)
	}
	defer unix.Close(fd)

	// Mark volumes. dirFds are kept open for the watch lifetime; closing any
	// of them drops its FAN_MARK_MOUNT mark.
	var dirFds []int
	defer func() {
		for _, d := range dirFds {
			unix.Close(d)
		}
	}()
	for _, vol := range f.volumes {
		supported, reason := classifyFeedVolume(vol)
		if !supported {
			logger.Info("UTMStack EDR: ransomware feed: skipping %s (%s)", vol, reason)
			continue
		}
		dirFd, err := unix.Open(vol, unix.O_RDONLY|unix.O_DIRECTORY, 0)
		if err != nil {
			logger.Error("UTMStack EDR: ransomware feed: open %s: %v", vol, err)
			continue
		}
		mask := effectiveMask(func(bit uint64) error {
			return unix.FanotifyMark(fd, unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT, bit, dirFd, "")
		})
		if mask == 0 {
			logger.Error("UTMStack EDR: ransomware feed: mark %s: no accepted bits", vol)
			unix.Close(dirFd)
			continue
		}
		if mask != fullFeedMask {
			logger.Info("UTMStack EDR: ransomware feed: %s marked 0x%x (degraded from 0x%x)", vol, mask, fullFeedMask)
		}
		dirFds = append(dirFds, dirFd)
	}
	if len(dirFds) == 0 {
		logger.Error("UTMStack EDR: ransomware feed: no volume could be marked, idle")
		<-ctx.Done()
		return ctx.Err()
	}
	logger.Info("UTMStack EDR: ransomware feed: started over %d volume(s)", len(dirFds))

	// Read loop: poll + non-blocking read so ctx cancellation is observed
	// promptly (a plain blocking read would hang the goroutine until an event).
	if err := unix.SetNonblock(fd, true); err != nil {
		return fmt.Errorf("setnonblock: %w", err)
	}
	var pfd [1]unix.PollFd
	pfd[0].Fd = int32(fd)
	pfd[0].Events = unix.POLLIN
	buf := make([]byte, fanReadBufSize)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		n, err := unix.Poll(pfd[:], 500)
		if err != nil {
			if errors.Is(err, syscall.EINTR) {
				continue
			}
			return fmt.Errorf("poll: %w", err)
		}
		if n == 0 {
			continue
		}
		got, rerr := unix.Read(fd, buf)
		if rerr != nil {
			if errors.Is(rerr, syscall.EAGAIN) || errors.Is(rerr, syscall.EWOULDBLOCK) {
				continue
			}
			return fmt.Errorf("read: %w", rerr)
		}
		if got <= 0 {
			continue
		}
		for _, meta := range parseFeedEvents(buf[:got]) {
			f.processFeedEvent(meta, sink)
		}
	}
}

// feedEvent is one decoded fanotify_event_metadata record. With
// FAN_REPORT_PIDFD (and no DFID_NAME) the record is the 24-byte header:
// Event_len u32, Vers u8, Reserved u8, Metadata_len u16, Mask u64, Fd i32,
// Pid i32. Parsed manually (little-endian) rather than via the x/sys struct
// because the kernel record is a byte stream, not a Go value.
type feedEvent struct {
	eventLen uint32
	mask     uint64
	fd       int32
	pid      int32
}

// parseFeedEvents walks a read buffer containing one or more fanotify
// records. Records are 8-byte aligned; each starts with the 24-byte header.
func parseFeedEvents(buf []byte) []feedEvent {
	var out []feedEvent
	off := 0
	for off < len(buf) {
		if off+24 > len(buf) {
			break
		}
		meta := feedEvent{
			eventLen: binary.LittleEndian.Uint32(buf[off:]),
			mask:     binary.LittleEndian.Uint64(buf[off+8:]),
			fd:       int32(binary.LittleEndian.Uint32(buf[off+16:])),
			pid:      int32(binary.LittleEndian.Uint32(buf[off+20:])),
		}
		if meta.eventLen < 24 || int(off)+int(meta.eventLen) > len(buf) {
			break
		}
		out = append(out, meta)
		off += (int(meta.eventLen) + 7) &^ 7
	}
	return out
}

// processFeedEvent resolves one event's path, maps its mask to a FileOp,
// delivers it to sink, and closes the event fd. The kernel opened the fd for
// us; leaking it exhausts fds under churn.
func (f fanotifyFeed) processFeedEvent(meta feedEvent, sink func(FileEvent)) {
	if meta.fd != unix.FAN_NOFD {
		defer unix.Close(int(meta.fd))
	}
	if meta.mask&unix.FAN_Q_OVERFLOW != 0 {
		logger.Error("UTMStack EDR: ransomware feed: fanotify queue overflow, events dropped")
		return
	}
	op, ok := opForMask(meta.mask)
	if !ok {
		return
	}
	if meta.fd == unix.FAN_NOFD {
		return
	}
	path, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", meta.fd))
	if err != nil {
		// Stale fd (file deleted between event and readlink); drop the event.
		return
	}
	sink(FileEvent{PID: int(meta.pid), Path: path, Op: op})
}

// effectiveMask marks each candidate bit on dirFd and returns the union of
// the bits the kernel accepted. tryBit performs one FAN_MARK_ADD for a single
// bit; fanotify OR-s accepted marks over the same object, so the result is the
// live mask. Bits are tried in the order of desiredFeedBits (writes first, so
// a kernel that accepts nothing still yields the write bits). If no bit is
// accepted the return is 0 and the caller skips the volume.
func effectiveMask(tryBit func(uint64) error) uint64 {
	var result uint64
	for _, bit := range desiredFeedBits {
		if tryBit(bit) == nil {
			result |= bit
		}
	}
	return result
}

// opForMask maps a fanotify event mask to a FileOp. FAN_CLOSE_WRITE and
// FAN_MODIFY both mean "a file was written" → OpWrite; FAN_CREATE → OpCreate;
// FAN_MOVED_TO → OpRename (the file arrived at a new path, the churn input);
// FAN_DELETE → OpDelete. FAN_Q_OVERFLOW is not a file op. Returns (0,false)
// for anything else. Write bits take priority so a compound mask still yields
// the op the sensors most need.
func opForMask(mask uint64) (FileOp, bool) {
	if mask&unix.FAN_Q_OVERFLOW != 0 {
		return 0, false
	}
	switch {
	case mask&unix.FAN_CLOSE_WRITE != 0:
		return OpWrite, true
	case mask&unix.FAN_MODIFY != 0:
		return OpWrite, true
	case mask&unix.FAN_CREATE != 0:
		return OpCreate, true
	case mask&unix.FAN_MOVED_TO != 0:
		return OpRename, true
	case mask&unix.FAN_DELETE != 0:
		return OpDelete, true
	default:
		return 0, false
	}
}

// classifyFeedVolume classifies the filesystem at path by its superblock
// magic and reports whether it is safe to mark, plus a short human label for
// logging. Network and virtual filesystems (NFS/9P, CIFS, overlay) are
// skipped — the same rule the main watcher applies.
func classifyFeedVolume(path string) (bool, string) {
	var sf unix.Statfs_t
	if err := unix.Statfs(path, &sf); err != nil {
		return false, fmt.Sprintf("statfs: %v", err)
	}
	switch sf.Type {
	case unix.NFS_SUPER_MAGIC: // NFS / 9P2000 (both 0x6969)
		return false, "nfs/9p"
	case unix.CIFS_SUPER_MAGIC:
		return false, "cifs"
	case unix.OVERLAYFS_SUPER_MAGIC:
		return false, "overlayfs"
	default:
		return true, fmt.Sprintf("magic 0x%x", sf.Type)
	}
}
