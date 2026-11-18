//go:build linux

package fanotify

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"golang.org/x/sys/unix"
)

type nativeBackend struct {
	fd              int
	permissionGroup bool
}

func kernelError(err error) error {
	if errors.Is(err, unix.EPERM) || errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EINVAL) || errors.Is(err, unix.EOPNOTSUPP) {
		return fmt.Errorf("%w: %v", ErrUnsupported, err)
	}
	return err
}
func smallFile(path string, maxBytes int) ([]byte, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, int64(maxBytes+1)))
	if e == nil && len(b) > maxBytes {
		e = errors.New("fanotify proc metadata limit")
	}
	return b, e
}
func fdMountID(fd int) (int, error) {
	b, e := smallFile(fmt.Sprintf("/proc/self/fdinfo/%d", fd), 4096)
	if e != nil {
		return 0, e
	}
	for _, line := range strings.Split(string(b), "\n") {
		if strings.HasPrefix(line, "mnt_id:") {
			return strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "mnt_id:")))
		}
	}
	return 0, errors.New("fanotify descriptor mount identity unavailable")
}
func mountPoints() (map[string]int, error) {
	b, e := smallFile("/proc/self/mountinfo", 1<<20)
	if e != nil {
		return nil, e
	}
	out := map[string]int{}
	replace := strings.NewReplacer(`\040`, " ", `\011`, "\t", `\012`, "\n", `\134`, `\`)
	for _, line := range strings.Split(string(b), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}
		id, err := strconv.Atoi(fields[0])
		if err != nil {
			return nil, err
		}
		out[replace.Replace(fields[4])] = id
	}
	return out, nil
}
func directoryFD(path string) (int, error) {
	return unix.Openat2(unix.AT_FDCWD, path, &unix.OpenHow{Flags: unix.O_RDONLY | unix.O_CLOEXEC | unix.O_DIRECTORY, Resolve: unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS})
}

// A Go process blocked in vfork can prevent its own GC safepoint while its
// child's exec waits for this monitor. Permission ownership must therefore stay
// off the daemon/executable-helper mounts. An independent permission broker is
// required before general executable-filesystem protection can be supported.
func executableMounts() (map[int]bool, error) {
	ids := map[int]bool{}
	fd, err := unix.Open("/proc/self/exe", unix.O_PATH|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	id, err := fdMountID(fd)
	unix.Close(fd)
	if err != nil || id <= 0 {
		return nil, errors.New("fanotify cannot pin the running executable mount")
	}
	ids[id] = true
	for _, path := range []string{"/bin", "/sbin", "/usr/bin", "/usr/sbin", "/usr/libexec", "/usr/local/bin", "/usr/local/sbin"} {
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return nil, err
		}
		fd, err := directoryFD(resolved)
		if err != nil {
			return nil, err
		}
		id, err := fdMountID(fd)
		unix.Close(fd)
		if err != nil || id <= 0 {
			return nil, errors.New("fanotify cannot pin a standard executable-helper mount")
		}
		ids[id] = true
	}
	return ids, nil
}
func openBackend(cfg Config) (backend, error) {
	return openNativeBackend(cfg, false)
}

// independentOwner is reserved for ServeBroker after it verifies its inherited
// parent channel. A Monitor sharing the scanner process must retain the guard.
func openNativeBackend(cfg Config, independentOwner bool) (backend, error) {
	mounts, e := mountPoints()
	if e != nil {
		return nil, e
	}
	declared := map[int]bool{}
	// Validate all scope and mount identities before activating any group.
	for _, path := range cfg.Mounts {
		want, ok := mounts[path]
		if !ok {
			return nil, fmt.Errorf("fanotify requires an explicit actual mountpoint: %q", path)
		}
		fd, err := directoryFD(path)
		if err != nil {
			return nil, kernelError(err)
		}
		id, err := fdMountID(fd)
		unix.Close(fd)
		if err != nil {
			return nil, err
		}
		if id != want {
			return nil, errors.New("fanotify mount changed during setup")
		}
		declared[id] = true
	}
	for _, path := range cfg.Roots {
		fd, err := directoryFD(path)
		if err != nil {
			return nil, kernelError(err)
		}
		id, err := fdMountID(fd)
		unix.Close(fd)
		if err != nil {
			return nil, err
		}
		if !declared[id] {
			return nil, fmt.Errorf("fanotify root is on an undeclared nested mount: %q", path)
		}
	}
	if cfg.Mode != ModeNotify && !independentOwner {
		executables, err := executableMounts()
		if err != nil {
			return nil, fmt.Errorf("%w: permission executable-mount safety: %v", ErrUnsupported, err)
		}
		for id := range declared {
			if executables[id] {
				return nil, fmt.Errorf("%w: Go permission scanning requires a dedicated data mount outside the EDR and executable-helper mounts", ErrUnsupported)
			}
		}
	}
	flags := uint(unix.FAN_CLOEXEC | unix.FAN_NONBLOCK | unix.FAN_REPORT_PIDFD)
	if cfg.Mode == ModeNotify {
		flags |= unix.FAN_CLASS_NOTIF
	} else {
		flags |= unix.FAN_CLASS_CONTENT
	}
	fd, e := unix.FanotifyInit(flags, unix.O_RDONLY|unix.O_LARGEFILE|unix.O_CLOEXEC|unix.O_NONBLOCK)
	if e != nil {
		return nil, kernelError(e)
	}
	source := &nativeBackend{fd: fd, permissionGroup: cfg.Mode != ModeNotify}
	ok := false
	defer func() {
		if !ok {
			source.Close()
		}
	}()
	// Permission groups own only their requested access gate. Inotify and the
	// separate ransomware notification group cover post-arrival modifications.
	var mask uint64
	if cfg.Mode == ModeNotify {
		mask = unix.FAN_MODIFY | unix.FAN_CLOSE_WRITE
	}
	if cfg.Mode == ModeExec {
		mask |= unix.FAN_OPEN_EXEC_PERM
	}
	if cfg.Mode == ModeOpen {
		mask |= unix.FAN_OPEN_PERM
	}
	for _, path := range cfg.Mounts {
		dir, err := directoryFD(path)
		if err != nil {
			return nil, err
		}
		id, err := fdMountID(dir)
		if err == nil && id != mounts[path] {
			err = errors.New("fanotify mount changed before mark")
		}
		if err == nil {
			err = unix.FanotifyMark(fd, unix.FAN_MARK_ADD|unix.FAN_MARK_MOUNT|unix.FAN_MARK_ONLYDIR, mask, dir, "")
		}
		unix.Close(dir)
		if err != nil {
			return nil, fmt.Errorf("fanotify mark %q: %w", path, kernelError(err))
		}
	}
	ok = true
	return source, nil
}
func (b *nativeBackend) Read(ctx context.Context) ([]rawEvent, error) {
	var buffer [8192]byte
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		polls := []unix.PollFd{{Fd: int32(b.fd), Events: unix.POLLIN}}
		n, e := unix.Poll(polls, 10)
		if errors.Is(e, unix.EINTR) {
			continue
		}
		if e != nil {
			return nil, e
		}
		if n == 0 {
			continue
		}
		if polls[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return nil, errors.New("fanotify group unavailable")
		}
		n, e = unix.Read(b.fd, buffer[:])
		if errors.Is(e, unix.EAGAIN) || errors.Is(e, unix.EINTR) {
			continue
		}
		if e != nil {
			return nil, e
		}
		if n == 0 {
			return nil, errors.New("fanotify empty read")
		}
		return decodeEvents(buffer[:n])
	}
}
func fileVersion(st *unix.Stat_t) string {
	return fmt.Sprintf("%d:%d:%d:%d:%d:%d:%d", st.Dev, st.Ino, st.Size, st.Mtim.Sec, st.Mtim.Nsec, st.Ctim.Sec, st.Ctim.Nsec)
}
func (b *nativeBackend) Inspect(raw rawEvent) (Event, error) {
	ev := Event{PID: raw.pid, Mask: raw.mask, Permission: raw.mask&(OpenPermission|OpenExecPermission) != 0}
	if raw.fd < 0 {
		return ev, errors.New("fanotify event has no pinned file descriptor")
	}
	var st unix.Stat_t
	if err := unix.Fstat(raw.fd, &st); err != nil {
		return ev, err
	}
	if st.Mode&unix.S_IFMT != unix.S_IFREG {
		return ev, errNotRegular
	}
	path, err := os.Readlink(fmt.Sprintf("/proc/self/fd/%d", raw.fd))
	if err != nil {
		return ev, err
	}
	if !strings.HasPrefix(path, "/") || len(path) > 4096 {
		return ev, errors.New("fanotify object has no stable scoped path")
	}
	ev.Path = path
	if strings.HasSuffix(path, " (deleted)") {
		return ev, errAmbiguousName
	}
	ev.Size = st.Size
	ev.Device, ev.Inode = uint64(st.Dev), st.Ino
	ev.version = fileVersion(&st)
	return ev, nil
}
func pidfdAlivePoll(fd int, poll func([]unix.PollFd, int) (int, error)) (bool, error) {
	if fd < 0 {
		return false, fmt.Errorf("kernel pidfd unavailable (%d)", fd)
	}
	p := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	// Signals (including Go runtime preemption) can interrupt poll even with
	// a zero timeout. An interrupted probe is not evidence of process exit.
	// Keep this metadata operation nonblocking and cap repeated interruptions.
	for attempt := 0; attempt < 8; attempt++ {
		n, err := poll(p, 0)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			return false, err
		}
		if n == 0 {
			return true, nil
		}
		if p[0].Revents&unix.POLLNVAL != 0 {
			return false, unix.EBADF
		}
		if n == 1 && p[0].Revents&(unix.POLLIN|unix.POLLHUP) != 0 {
			return false, nil
		}
		return false, fmt.Errorf("unexpected pidfd readiness: count=%d events=%#x", n, p[0].Revents)
	}
	return false, unix.EINTR
}

func attributionError(pid int, stage string, err error) error {
	if err != nil {
		return fmt.Errorf("%w: pid=%d %s: %w", ErrAttribution, pid, stage, err)
	}
	return fmt.Errorf("%w: pid=%d %s", ErrAttribution, pid, stage)
}
func (b *nativeBackend) Attribute(raw rawEvent, ev *Event) error {
	if raw.pid <= 1 {
		return attributionError(raw.pid, "invalid event process ID", nil)
	}
	alive, err := pidfdAlivePoll(raw.pidfd, unix.Poll)
	if err != nil || !alive {
		return attributionError(raw.pid, "pidfd not live before metadata", err)
	}
	// An exec permission request can hold the origin's exec/mm locks while it
	// waits for our reply. Reading /proc/PID/exe or cmdline here would wait for
	// those same locks and deadlock the sole permission owner. Identity-only
	// stat/pidfd checks suffice; process-image metadata is intentionally absent.
	// This also applies to earlier notifications on a permission group, which
	// could otherwise deadlock behind an exec permission already queued to us.
	before, e := (linuxproc.Reader{}).Read(raw.pid, !b.permissionGroup && !ev.Permission)
	if e != nil {
		return attributionError(raw.pid, "initial process stat", e)
	}
	if before.StartTicks <= 0 {
		return attributionError(raw.pid, "initial process identity missing", nil)
	}
	info, e := smallFile(fmt.Sprintf("/proc/self/fdinfo/%d", raw.pidfd), 4096)
	if e != nil {
		return attributionError(raw.pid, "pidfd metadata", e)
	}
	reported := -1
	for _, line := range strings.Split(string(info), "\n") {
		if strings.HasPrefix(line, "Pid:") {
			reported, _ = strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pid:")))
		}
	}
	after, e := (linuxproc.Reader{}).Read(raw.pid, false)
	if e != nil {
		return attributionError(raw.pid, "final process stat", e)
	}
	if reported != raw.pid {
		return attributionError(raw.pid, fmt.Sprintf("pidfd reports different process ID %d", reported), nil)
	}
	if after.StartTicks != before.StartTicks {
		return attributionError(raw.pid, "process start identity changed", nil)
	}
	alive, err = pidfdAlivePoll(raw.pidfd, unix.Poll)
	if err != nil || !alive {
		return attributionError(raw.pid, "pidfd not live after metadata", err)
	}
	ev.OSIdentity = before.StartTicks
	ev.PPID = before.PPID
	ev.Image = before.Image
	ev.Cmdline = before.Cmdline
	return nil
}
func (b *nativeBackend) Unchanged(raw rawEvent, ev Event) error {
	var st unix.Stat_t
	if e := unix.Fstat(raw.fd, &st); e != nil {
		return e
	}
	if fileVersion(&st) != ev.version {
		return errors.New("fanotify file changed during analysis")
	}
	return nil
}
func (b *nativeBackend) Duplicate(raw rawEvent) (*os.File, error) {
	fd, e := unix.FcntlInt(uintptr(raw.fd), unix.F_DUPFD_CLOEXEC, 0)
	if e != nil {
		return nil, e
	}
	return os.NewFile(uintptr(fd), "fanotify-pinned-input"), nil
}
func (b *nativeBackend) Respond(raw rawEvent, allow bool) error {
	var response [8]byte
	binary.NativeEndian.PutUint32(response[:], uint32(raw.fd))
	decision := uint32(unix.FAN_DENY)
	if allow {
		decision = unix.FAN_ALLOW
	}
	binary.NativeEndian.PutUint32(response[4:], decision)
	deadline := time.Now().Add(20 * time.Millisecond)
	for {
		n, e := unix.Write(b.fd, response[:])
		if errors.Is(e, unix.EINTR) {
			continue
		}
		if errors.Is(e, unix.EAGAIN) && time.Now().Before(deadline) {
			continue
		}
		if e != nil {
			return e
		}
		if n != len(response) {
			return errors.New("fanotify partial permission response")
		}
		return nil
	}
}
func (b *nativeBackend) Release(raw rawEvent) {
	seen := map[int]bool{}
	for _, fd := range append([]int{raw.fd, raw.pidfd}, raw.pidfds...) {
		if fd >= 0 && !seen[fd] {
			unix.Close(fd)
			seen[fd] = true
		}
	}
}
func (b *nativeBackend) Close() error {
	if b.fd < 0 {
		return nil
	}
	fd := b.fd
	b.fd = -1
	return unix.Close(fd)
}
