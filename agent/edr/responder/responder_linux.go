//go:build linux

package responder

import (
	"context"
	"errors"
	"fmt"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/linuxproc"
	"golang.org/x/sys/unix"
	"os"
	"sync"
	"time"
)

type OSKiller struct{}

var _ Identifier = OSKiller{}

func processError(err error) error {
	if errors.Is(err, os.ErrNotExist) || errors.Is(err, unix.ESRCH) {
		return ErrProcessGone
	}
	if errors.Is(err, unix.ENOSYS) || errors.Is(err, unix.EOPNOTSUPP) {
		return fmt.Errorf("%w: %v", ErrUnsupported, err)
	}
	return err
}
func protectedPID(pid int) bool { return pid <= 1 || pid == os.Getpid() }
func (OSKiller) ProcessIdentity(pid int) (int64, error) {
	p, err := (linuxproc.Reader{}).Read(pid, false)
	if err != nil {
		return 0, processError(err)
	}
	return p.StartTicks, nil
}
func (OSKiller) KillPID(pid int) error {
	id, err := (OSKiller{}).ProcessIdentity(pid)
	if err != nil {
		return err
	}
	return (OSKiller{}).KillIdentity(pid, id)
}
func (OSKiller) KillIdentity(pid int, id int64) error {
	fd, err := identityFD(pid, id)
	if err != nil {
		return err
	}
	defer unix.Close(fd)
	return processError(unix.PidfdSendSignal(fd, unix.SIGKILL, nil, 0))
}

// There is deliberately no kill(2) fallback: a kernel without pidfds cannot
// provide the check-and-signal lifetime guarantee this implementation requires.
func identityFD(pid int, id int64) (int, error) {
	if protectedPID(pid) || id <= 0 {
		return -1, errors.New("refusing response to protected or unidentified process")
	}
	fd, err := unix.PidfdOpen(pid, 0)
	if err != nil {
		return -1, processError(err)
	}
	current, err := (OSKiller{}).ProcessIdentity(pid)
	if err != nil {
		unix.Close(fd)
		return -1, err
	}
	if current != id {
		unix.Close(fd)
		return -1, errPIDReused
	}
	return fd, nil
}
func LiveProcess(pid int) (string, int, error) {
	p, err := (linuxproc.Reader{}).Read(pid, true)
	if err != nil {
		return "", 0, processError(err)
	}
	return p.Image, p.PPID, nil
}

func linuxTree(root int, identity int64) ([]linuxproc.Process, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	all, err := (linuxproc.Reader{}).Snapshot(ctx, false)
	if err != nil {
		return nil, err
	}
	byPID := make(map[int]linuxproc.Process, len(all))
	children := map[int][]int{}
	for _, p := range all {
		byPID[p.PID] = p
		children[p.PPID] = append(children[p.PPID], p.PID)
	}
	p, ok := byPID[root]
	if !ok {
		return nil, ErrProcessGone
	}
	if p.StartTicks != identity {
		return nil, errPIDReused
	}
	var order []linuxproc.Process
	seen := map[int]bool{}
	var visit func(int)
	visit = func(pid int) {
		if seen[pid] {
			return
		}
		seen[pid] = true
		p := byPID[pid]
		for _, child := range children[pid] {
			if byPID[child].StartTicks >= p.StartTicks {
				visit(child)
			}
		}
		order = append(order, p)
	}
	visit(root)
	return order, nil
}
func liveProcessTree(root int) ([]processParent, bool) {
	id, err := (OSKiller{}).ProcessIdentity(root)
	if err != nil {
		return nil, false
	}
	tree, err := linuxTree(root, id)
	if err != nil {
		return nil, false
	}
	var pids []processParent
	for _, p := range tree {
		pids = append(pids, processParent{p.PID, p.PPID})
	}
	return pids, true
}

// KillTreeIdentity binds a delayed launch verdict to the originally observed
// root and every child's captured creation identity; stale proctable PIDs are
// not unioned into this native snapshot. Descendants can still escape/reparent
// between observations; this is after-the-fact containment, not a process gate.
func (r *Responder) KillTreeIdentity(root int, identity int64, signature string) ([]int, error) {
	if protectedPID(root) {
		return nil, errors.New("refusing response to protected process")
	}
	idf, ok := r.kill.(Identifier)
	if !ok {
		return nil, ErrUnsupported
	}
	tree, err := linuxTree(root, identity)
	if err != nil {
		return nil, err
	}
	current, err := idf.ProcessIdentity(root)
	if err != nil {
		return nil, err
	}
	if current != identity {
		return nil, errPIDReused
	}
	var killed []int
	var errs []error
	for _, p := range tree {
		if protectedPID(p.PID) {
			errs = append(errs, errors.New("protected descendant skipped"))
			continue
		}
		key := processKey{pid: p.PID, id: p.StartTicks, src: identityOS}
		if !r.claims.take(key, r.now()) {
			continue
		}
		if err := idf.KillIdentity(p.PID, p.StartTicks); err != nil {
			r.claims.release(key)
			if !errors.Is(err, ErrProcessGone) {
				errs = append(errs, err)
			}
			continue
		}
		killed = append(killed, p.PID)
		meta, _ := r.table.Get(p.PID)
		if r.spool != nil {
			ev := event.NewProcessAction(event.ActionKilled, event.SourceProcessWatcher, event.ProcInfo{PID: p.PID, PPID: p.PPID, Image: meta.Image, Cmdline: meta.Cmdline}, "malicious", signature)
			if js, e := ev.ToJSON(); e == nil {
				_ = r.spool.Append(js)
			}
		}
		r.table.RemoveObserved(p.PID, p.StartTicks)
	}
	return killed, errors.Join(errs...)
}

type heldProcess struct {
	fd       int
	identity int64
	owners   int
}

var suspendedLinux = struct {
	sync.Mutex
	closed bool
	byPID  map[int]heldProcess
}{byPID: map[int]heldProcess{}}

type OSSuspender struct{}

func (OSSuspender) SuspendPID(pid int) error {
	id, e := (OSKiller{}).ProcessIdentity(pid)
	if e != nil {
		return e
	}
	return (OSSuspender{}).SuspendIdentity(pid, id)
}
func (OSSuspender) SuspendIdentity(pid int, id int64) error {
	suspendedLinux.Lock()
	defer suspendedLinux.Unlock()
	if suspendedLinux.closed {
		return ErrSuspensionsClosed
	}
	for old, held := range suspendedLinux.byPID {
		if e := unix.PidfdSendSignal(held.fd, 0, nil, 0); errors.Is(e, unix.ESRCH) {
			unix.Close(held.fd)
			delete(suspendedLinux.byPID, old)
		}
	}
	if held, ok := suspendedLinux.byPID[pid]; ok {
		if held.identity != id {
			return errPIDReused
		}
		if held.owners >= 1024 {
			return errors.New("suspension owner limit")
		}
		held.owners++
		suspendedLinux.byPID[pid] = held
		return nil
	}
	if len(suspendedLinux.byPID) >= 1024 {
		return errors.New("suspension descriptor limit")
	}
	fd, e := identityFD(pid, id)
	if e != nil {
		return e
	}
	p, e := (linuxproc.Reader{}).Read(pid, false)
	if e != nil {
		unix.Close(fd)
		return e
	}
	if p.State == 'T' || p.State == 't' {
		unix.Close(fd)
		return errors.New("process already stopped by another owner")
	}
	if e = unix.PidfdSendSignal(fd, unix.SIGSTOP, nil, 0); e != nil {
		unix.Close(fd)
		return processError(e)
	}
	suspendedLinux.byPID[pid] = heldProcess{fd: fd, identity: id, owners: 1}
	return nil
}
func (OSSuspender) ResumePID(pid int) error { return (OSSuspender{}).ResumeIdentity(pid, 0) }
func (OSSuspender) ResumeIdentity(pid int, id int64) error {
	suspendedLinux.Lock()
	defer suspendedLinux.Unlock()
	held, ok := suspendedLinux.byPID[pid]
	if !ok {
		return errors.New("process was not suspended by this service")
	}
	if id != 0 && held.identity != id {
		return errPIDReused
	}
	if held.owners > 1 {
		held.owners--
		suspendedLinux.byPID[pid] = held
		return nil
	}
	err := processError(unix.PidfdSendSignal(held.fd, unix.SIGCONT, nil, 0))
	unix.Close(held.fd)
	delete(suspendedLinux.byPID, pid)
	return err
}

// CloseSuspensions resumes only pidfd-bound processes this service stopped.
// The owner should call this during shutdown; it never signals a reused PID.
func CloseSuspensions() error {
	suspendedLinux.Lock()
	defer suspendedLinux.Unlock()
	// Seal before releasing handles: guards already admitted at shutdown must
	// never add a new owned suspension after this cleanup returns.
	suspendedLinux.closed = true
	var errs []error
	for pid, p := range suspendedLinux.byPID {
		if err := processError(unix.PidfdSendSignal(p.fd, unix.SIGCONT, nil, 0)); err != nil && !errors.Is(err, ErrProcessGone) {
			errs = append(errs, err)
		}
		unix.Close(p.fd)
		delete(suspendedLinux.byPID, pid)
	}
	return errors.Join(errs...)
}
