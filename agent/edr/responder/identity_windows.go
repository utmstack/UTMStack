//go:build windows

package responder

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"golang.org/x/sys/windows"
)

type windowsProcess struct {
	pid, ppid int
	identity  int64
}

// Read creation time and parent from the same process handle. Toolhelp PID/PPID
// alone may describe a prior occupant of a PID by the time it is inspected.
func processObservation(pid int) (windowsProcess, error) {
	h, e := openProcess(windows.PROCESS_QUERY_INFORMATION, pid)
	if e != nil {
		return windowsProcess{}, e
	}
	defer windows.CloseHandle(h)
	id, e := creationTime(h, pid)
	if e != nil {
		return windowsProcess{}, e
	}
	var basic windows.PROCESS_BASIC_INFORMATION
	var n uint32
	if e = windows.NtQueryInformationProcess(h, windows.ProcessBasicInformation, unsafe.Pointer(&basic), uint32(unsafe.Sizeof(basic)), &n); e != nil {
		return windowsProcess{}, e
	}
	if int(basic.UniqueProcessId) != pid || id <= 0 {
		return windowsProcess{}, errPIDReused
	}
	return windowsProcess{pid, int(basic.InheritedFromUniqueProcessId), id}, nil
}

func windowsIdentityTree(root int, identity int64) ([]windowsProcess, error) {
	original, e := processObservation(root)
	if e != nil {
		return nil, e
	}
	if original.identity != identity {
		return nil, errPIDReused
	}
	snap, e := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if e != nil {
		return nil, e
	}
	defer windows.CloseHandle(snap)
	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	if e = windows.Process32First(snap, &entry); e != nil {
		return nil, e
	}
	byPID := map[int]windowsProcess{root: original}
	children := map[int][]int{}
	deadline := time.Now().Add(2 * time.Second)
	for count := 0; ; count++ {
		if count >= 32768 || time.Now().After(deadline) {
			return nil, errors.New("process tree snapshot limit")
		}
		pid := int(entry.ProcessID)
		if !protectedPID(pid) {
			if p, err := processObservation(pid); err == nil {
				byPID[pid] = p
				children[p.ppid] = append(children[p.ppid], pid)
			}
		}
		if e = windows.Process32Next(snap, &entry); e != nil {
			if !errors.Is(e, windows.ERROR_NO_MORE_FILES) {
				return nil, e
			}
			break
		}
	}
	current, e := processObservation(root)
	if e != nil {
		return nil, e
	}
	if current.identity != identity {
		return nil, errPIDReused
	}
	var out []windowsProcess
	seen := map[int]bool{}
	var visit func(int)
	visit = func(pid int) {
		if seen[pid] {
			return
		}
		seen[pid] = true
		p := byPID[pid]
		for _, child := range children[pid] {
			if byPID[child].identity >= p.identity {
				visit(child)
			}
		}
		out = append(out, p)
	}
	visit(root)
	return out, nil
}

// KillTreeIdentity applies a delayed verdict only to the original root and
// captured child instances. It deliberately excludes stale proctable-only PIDs.
func (r *Responder) KillTreeIdentity(root int, identity int64, signature string) ([]int, error) {
	idf, ok := r.kill.(Identifier)
	if !ok {
		return nil, ErrUnsupported
	}
	tree, e := windowsIdentityTree(root, identity)
	if e != nil {
		return nil, e
	}
	current, e := idf.ProcessIdentity(root)
	if e != nil {
		return nil, e
	}
	if current != identity {
		return nil, errPIDReused
	}
	var killed []int
	var errs []error
	for _, p := range tree {
		key := processKey{pid: p.pid, id: p.identity, src: identityOS}
		if !r.claims.take(key, r.now()) {
			continue
		}
		if e = idf.KillIdentity(p.pid, p.identity); e != nil {
			r.claims.release(key)
			if !errors.Is(e, ErrProcessGone) {
				errs = append(errs, e)
			}
			continue
		}
		killed = append(killed, p.pid)
		meta, _ := r.table.Get(p.pid)
		if r.spool != nil {
			ev := event.NewProcessAction(event.ActionKilled, event.SourceProcessWatcher, event.ProcInfo{PID: p.pid, PPID: p.ppid, Image: meta.Image, Cmdline: meta.Cmdline}, "malicious", signature)
			if js, err := ev.ToJSON(); err == nil {
				_ = r.spool.Append(js)
			}
		}
		r.table.RemoveObserved(p.pid, p.identity)
	}
	return killed, errors.Join(errs...)
}

type heldWindowsProcess struct {
	handle   windows.Handle
	identity int64
	owners   int
}

var suspendedWindows = struct {
	sync.Mutex
	closed bool
	byPID  map[int]heldWindowsProcess
}{byPID: map[int]heldWindowsProcess{}}

func (OSSuspender) SuspendPID(pid int) error {
	id, e := (OSKiller{}).ProcessIdentity(pid)
	if e != nil {
		return e
	}
	return (OSSuspender{}).SuspendIdentity(pid, id)
}
func (OSSuspender) SuspendIdentity(pid int, id int64) error {
	if id <= 0 || protectedPID(pid) {
		return errors.New("refusing suspension of protected or unidentified process")
	}
	suspendedWindows.Lock()
	defer suspendedWindows.Unlock()
	if suspendedWindows.closed {
		return ErrSuspensionsClosed
	}
	for old, held := range suspendedWindows.byPID {
		var code uint32
		if e := windows.GetExitCodeProcess(held.handle, &code); e == nil && code != 259 {
			windows.CloseHandle(held.handle)
			delete(suspendedWindows.byPID, old)
		}
	}
	if held, ok := suspendedWindows.byPID[pid]; ok {
		if held.identity != id {
			return errPIDReused
		}
		if held.owners >= 1024 {
			return errors.New("suspension owner limit")
		}
		held.owners++
		suspendedWindows.byPID[pid] = held
		return nil
	}
	if len(suspendedWindows.byPID) >= 1024 {
		return errors.New("suspension handle limit")
	}
	h, e := openProcess(processSuspendResume|windows.PROCESS_QUERY_LIMITED_INFORMATION, pid)
	if e != nil {
		return e
	}
	created, e := creationTime(h, pid)
	if e != nil || created != id {
		windows.CloseHandle(h)
		if e != nil {
			return e
		}
		return errPIDReused
	}
	if e = ntProcessCall(h, procNtSuspendProc); e != nil {
		windows.CloseHandle(h)
		return e
	}
	suspendedWindows.byPID[pid] = heldWindowsProcess{h, id, 1}
	return nil
}
func ntProcessCall(h windows.Handle, proc *windows.LazyProc) error {
	status, _, _ := proc.Call(uintptr(h))
	if status != 0 {
		return fmt.Errorf("nt process call failed: status 0x%x", status)
	}
	return nil
}
func (OSSuspender) ResumePID(pid int) error { return (OSSuspender{}).ResumeIdentity(pid, 0) }
func (OSSuspender) ResumeIdentity(pid int, id int64) error {
	suspendedWindows.Lock()
	defer suspendedWindows.Unlock()
	held, ok := suspendedWindows.byPID[pid]
	if !ok {
		return errors.New("process was not suspended by this service")
	}
	if id != 0 && held.identity != id {
		return errPIDReused
	}
	if held.owners > 1 {
		held.owners--
		suspendedWindows.byPID[pid] = held
		return nil
	}
	e := ntProcessCall(held.handle, procNtResumeProc)
	windows.CloseHandle(held.handle)
	delete(suspendedWindows.byPID, pid)
	return e
}

// CloseSuspensions balances this service's native suspension on the original
// process handles; it never opens the current occupant of an old numeric PID.
func CloseSuspensions() error {
	suspendedWindows.Lock()
	defer suspendedWindows.Unlock()
	// Seal before releasing handles: guards already admitted at shutdown must
	// never add a new owned suspension after this cleanup returns.
	suspendedWindows.closed = true
	var errs []error
	for pid, p := range suspendedWindows.byPID {
		if e := ntProcessCall(p.handle, procNtResumeProc); e != nil {
			errs = append(errs, e)
		}
		windows.CloseHandle(p.handle)
		delete(suspendedWindows.byPID, pid)
	}
	return errors.Join(errs...)
}
