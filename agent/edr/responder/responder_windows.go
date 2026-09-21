//go:build windows

package responder

import (
	"errors"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const processSuspendResume = 0x0800

// liveDescendants returns rootPID and all of its transitive children from a live
// OS process snapshot (Toolhelp), ordered leaves-first. This catches children the
// async WMI-fed proctable has not registered yet — the common tree-kill race where
// an encryptor spawns a helper microseconds before it is terminated. rootPID is
// alive at snapshot time, so PPID==rootPID reliably means a real child (the classic
// stale-PPID reuse hazard only applies to already-dead parents). Returns ok=false
// if the snapshot can't be taken, so the caller falls back to the proctable.
func liveDescendants(rootPID int) ([]int, bool) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, false
	}
	defer windows.CloseHandle(snap)

	children := map[int][]int{}
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	if err := windows.Process32First(snap, &e); err != nil {
		return nil, false
	}
	for {
		children[int(e.ParentProcessID)] = append(children[int(e.ParentProcessID)], int(e.ProcessID))
		if err := windows.Process32Next(snap, &e); err != nil {
			break
		}
	}

	var order []int
	seen := map[int]bool{}
	var visit func(int)
	visit = func(id int) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, c := range children[id] {
			if c != id { // guard against a self-parent cycle from PID reuse
				visit(c)
			}
		}
		order = append(order, id) // post-order => leaves first
	}
	visit(rootPID)
	return order, true
}

type OSKiller struct{}

func (OSKiller) KillPID(pid int) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		// A PID that no longer exists yields ERROR_INVALID_PARAMETER — the process
		// was already terminated (benign tree-kill race).
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return ErrProcessGone
		}
		return fmt.Errorf("open pid %d: %w", pid, err)
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}

type OSSuspender struct{}

var (
	ntdll             = windows.NewLazySystemDLL("ntdll.dll")
	procNtSuspendProc = ntdll.NewProc("NtSuspendProcess")
	procNtResumeProc  = ntdll.NewProc("NtResumeProcess")
)

func (OSSuspender) SuspendPID(pid int) error { return suspendResume(pid, procNtSuspendProc) }
func (OSSuspender) ResumePID(pid int) error  { return suspendResume(pid, procNtResumeProc) }

func suspendResume(pid int, proc *windows.LazyProc) error {
	h, err := windows.OpenProcess(processSuspendResume, false, uint32(pid))
	if err != nil {
		return fmt.Errorf("open pid %d: %w", pid, err)
	}
	defer windows.CloseHandle(h)
	r, _, _ := proc.Call(uintptr(h))
	if r != 0 { // NTSTATUS != STATUS_SUCCESS
		return fmt.Errorf("nt call failed: status 0x%x", r)
	}
	return nil
}
