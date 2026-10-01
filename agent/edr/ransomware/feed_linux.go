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

// feedMask is the set of events marked on every volume. Probed against the
// target kernel (FAN_MARK_MOUNT): FAN_MODIFY and FAN_CLOSE_WRITE are accepted;
// FAN_DELETE, FAN_DELETE_SELF, FAN_RENAME, FAN_CREATE and FAN_MOVED_TO are
// rejected with EINVAL at mount level. A file write generates both FAN_MODIFY
// and FAN_CLOSE_WRITE, so the mask covers "a file was written" with clean
// events. Delete/rename-based tampering is invisible to this feed and is
// covered by the T1490 command rules instead (see rules_linux.go).
const feedMask = uint64(unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE)

// fanReadBufSize: a single read from the fanotify fd returns one or more
// records; 64 KiB matches the watcher and is enough for a batch of events.
const fanReadBufSize = 64 * 1024

// fanotifyFeed is the Linux fanotify file-activity feed. It marks each
// configured volume at mount level (FAN_MARK_MOUNT) for FAN_MODIFY |
// FAN_CLOSE_WRITE and delivers per-process FileEvents (PID from
// FAN_REPORT_PIDFD, path from readlink of the event fd) to the guard's sink.
// It is observational only (no FAN_CLASS_CONTENT): the main watcher already
// owns permission mode.
type fanotifyFeed struct {
	volumes []string
}

// NewFeed returns the Linux fanotify file-activity feed over the given
// volumes. With no volumes Run blocks until ctx is cancelled (idle).
func NewFeed(vols ...string) FileActivityFeed { return fanotifyFeed{volumes: vols} }

// Run starts the fanotify feed and blocks until ctx is cancelled.
//
// Setup:
//  1. fanotify_init with FAN_CLOEXEC|FAN_REPORT_PIDFD (PID in metadata.pid).
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

	fd, err := unix.FanotifyInit(unix.FAN_CLOEXEC|unix.FAN_REPORT_PIDFD, 0)
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
		if err := unix.FanotifyMark(fd, unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT, feedMask, dirFd, ""); err != nil {
			logger.Error("UTMStack EDR: ransomware feed: mark %s: %v", vol, err)
			unix.Close(dirFd)
			continue
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

// opForMask maps a fanotify event mask to a FileOp. Both FAN_CLOSE_WRITE and
// FAN_MODIFY mean "a file was written" → OpWrite (the only ops in feedMask).
// FAN_Q_OVERFLOW is not a file op. Returns (0,false) for anything else.
func opForMask(mask uint64) (FileOp, bool) {
	if mask&unix.FAN_Q_OVERFLOW != 0 {
		return 0, false
	}
	switch {
	case mask&unix.FAN_CLOSE_WRITE != 0:
		return OpWrite, true
	case mask&unix.FAN_MODIFY != 0:
		return OpWrite, true
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
