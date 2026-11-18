//go:build linux

package watcher

import (
	"context"
	"errors"
	"fmt"
	"github.com/utmstack/UTMStack/shared/logger"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const maxLinuxWatches = 8192
const maxLinuxReconcileFiles = 65536
const maxLinuxReconcileEntries = 131072
const maxLinuxWatchDepth = 64

// Admission feedback is optional for existing watcher sinks. The Linux service
// implements it so a full queue cannot silently discard persistent work.
type admissionSink interface {
	TryEnqueue(path, op string) bool
}

var errLinuxAdmission = fmt.Errorf("%w: scan queue admission refused; pending work retained", ErrResync)

func (w *Watcher) tryEnqueue(path, op string) bool {
	if sink, ok := w.sink.(admissionSink); ok {
		return sink.TryEnqueue(path, op)
	}
	w.sink.Enqueue(path, op)
	return true
}

func (w *Watcher) enqueueReconcile(ctx context.Context, path string) error {
	return w.enqueueWait(ctx, path, "reconcile")
}

func (w *Watcher) enqueueWait(ctx context.Context, path, op string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if w.tryEnqueue(path, op) {
			return nil
		}
		w.setHealth(errLinuxAdmission)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(20 * time.Millisecond):
		}
	}
}

type Watcher struct {
	roots    []string
	sink     Sink
	excluded func(string) bool
	mu       sync.Mutex
	health   error
}

func New(roots []string, sink Sink, _ CursorStore, excluded func(string) bool) *Watcher {
	if excluded == nil {
		excluded = func(string) bool { return false }
	}
	return &Watcher{roots: append([]string(nil), roots...), sink: sink, excluded: excluded, health: errors.New("file monitor not started")}
}
func (w *Watcher) Health() error       { w.mu.Lock(); defer w.mu.Unlock(); return w.health }
func (w *Watcher) setHealth(err error) { w.mu.Lock(); w.health = err; w.mu.Unlock() }
func (w *Watcher) Run(ctx context.Context) {
	for {
		err := w.RunError(ctx)
		if ctx.Err() != nil {
			return
		}
		logger.Error("UTMStack EDR file monitor unavailable: %v; reconciling", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(3 * time.Second):
		}
	}
}

type inotifyScope struct {
	w       *Watcher
	fd      int
	paths   map[int32]string
	files   int
	entries int
}

func (s *inotifyScope) addTree(ctx context.Context, root string) error {
	// WalkDir reads/sorts an entire directory before the callback can enforce a
	// budget. Stream small batches instead, retaining at most one per depth.
	var walk func(string, int) error
	walk = func(path string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if s.entries >= maxLinuxReconcileEntries {
			return ErrWatchLimit
		}
		s.entries++
		if linuxPseudoPath(path) || s.w.excluded(path) {
			return nil
		}
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			if path == root {
				return fmt.Errorf("%w: symlink root", ErrUnavailable)
			}
			return nil
		}
		if depth > maxLinuxWatchDepth {
			return ErrWatchLimit
		}
		if !info.IsDir() {
			if info.Mode().IsRegular() {
				if s.files >= maxLinuxReconcileFiles {
					return ErrWatchLimit
				}
				s.files++
				if err := s.w.enqueueReconcile(ctx, path); err != nil {
					return err
				}
			}
			return nil
		}
		if len(s.paths) >= maxLinuxWatches {
			return ErrWatchLimit
		}
		wd, err := unix.InotifyAddWatch(s.fd, path, unix.IN_ONLYDIR|unix.IN_DONT_FOLLOW|unix.IN_CREATE|unix.IN_CLOSE_WRITE|unix.IN_MOVED_TO|unix.IN_MOVED_FROM|unix.IN_DELETE_SELF|unix.IN_MOVE_SELF)
		if err != nil {
			return err
		}
		s.paths[int32(wd)] = path
		// Refuse a directory swapped for a final-component symlink between lstat
		// and open. Ancestor replacement remains a documented path-based limit.
		dirFD, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_DIRECTORY|unix.O_NOFOLLOW, 0)
		if err != nil {
			return err
		}
		dir := os.NewFile(uintptr(dirFD), path)
		defer dir.Close()
		for {
			if err := ctx.Err(); err != nil {
				return err
			}
			entries, readErr := dir.ReadDir(128)
			for _, entry := range entries {
				if err := walk(filepath.Join(path, entry.Name()), depth+1); err != nil {
					return err
				}
			}
			if readErr == io.EOF {
				return nil
			}
			if readErr != nil {
				return readErr
			}
		}
	}
	return walk(root, 0)
}

// Inotify is scoped and volatile, unlike NTFS USN. Startup/recovery and periodic
// reconciliation cover existing files, but this does not promise every transient
// write or exec event, durable replay, or pre-execution prevention.
func (w *Watcher) RunError(ctx context.Context) (err error) {
	defer func() { w.setHealth(err) }()
	if w.sink == nil || len(w.roots) == 0 {
		return fmt.Errorf("%w: explicit absolute watched paths required", ErrUnavailable)
	}
	fd, err := unix.InotifyInit1(unix.IN_NONBLOCK | unix.IN_CLOEXEC)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	s := inotifyScope{w: w, fd: fd, paths: make(map[int32]string)}
	for _, root := range w.roots {
		if !filepath.IsAbs(root) || filepath.Clean(root) != root || linuxPseudoPath(root) {
			return fmt.Errorf("%w: invalid watched path %q", ErrUnavailable, root)
		}
		if err := s.addTree(ctx, root); err != nil {
			return err
		}
	}
	if len(s.paths) == 0 {
		return fmt.Errorf("%w: no readable directories watched", ErrUnavailable)
	}
	w.setHealth(nil)
	buf := make([]byte, 64<<10)
	deadline := time.Now().Add(5 * time.Minute)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return ErrResync
		}
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		n, e := unix.Poll(poll, 250)
		if errors.Is(e, unix.EINTR) {
			continue
		}
		if e != nil {
			return e
		}
		if n == 0 {
			w.setHealth(nil) // The bounded user batch and kernel backlog are drained.
			continue
		}
		if poll[0].Revents&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return ErrResync
		}
		n, e = unix.Read(fd, buf)
		if errors.Is(e, unix.EAGAIN) || errors.Is(e, unix.EINTR) {
			continue
		}
		if e != nil {
			return e
		}
		if n == 0 {
			return ErrResync
		}
		events, e := decodeLinuxEvents(buf[:n])
		if e != nil {
			return e
		}
		for _, ev := range events {
			if ev.mask&unix.IN_Q_OVERFLOW != 0 {
				return ErrResync
			}
			dir, ok := s.paths[ev.wd]
			if !ok {
				continue
			}
			if ev.mask&(unix.IN_DELETE_SELF|unix.IN_MOVE_SELF|unix.IN_IGNORED) != 0 {
				return ErrResync
			}
			if ev.name == "" {
				continue
			}
			path := filepath.Join(dir, ev.name)
			if w.excluded(path) || linuxPseudoPath(path) {
				continue
			}
			if ev.mask&unix.IN_ISDIR != 0 {
				if ev.mask&(unix.IN_MOVED_FROM|unix.IN_MOVED_TO) != 0 {
					return ErrResync
				}
				if ev.mask&unix.IN_CREATE != 0 {
					if e := s.addTree(ctx, path); e != nil {
						if errors.Is(e, os.ErrNotExist) {
							continue
						}
						return e
					}
				}
				continue
			}
			if ev.mask&(unix.IN_CLOSE_WRITE|unix.IN_MOVED_TO|unix.IN_CREATE) != 0 {
				info, e := os.Lstat(path)
				if e == nil && info.Mode().IsRegular() {
					if err := w.enqueueWait(ctx, path, "inotify"); err != nil {
						return err
					}
				}
			}
		}
	}
}
