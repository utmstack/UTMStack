//go:build windows

package responder

import (
	"errors"
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

const processSuspendResume = 0x0800

// liveProcessTree returns rootPID and all of its transitive children from a live
// OS process snapshot (Toolhelp), ordered leaves-first. This catches children the
// async WMI-fed proctable has not registered yet — the common tree-kill race where
// an encryptor spawns a helper microseconds before it is terminated. rootPID is
// alive at snapshot time, so PPID==rootPID reliably means a real child (the classic
// stale-PPID reuse hazard only applies to already-dead parents). Returns ok=false
// if the snapshot can't be taken, so the caller falls back to the proctable.
func liveProcessTree(rootPID int) ([]processParent, bool) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, false
	}
	defer windows.CloseHandle(snap)

	children := map[int][]int{}
	parents := map[int]int{}
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	if err := windows.Process32First(snap, &e); err != nil {
		return nil, false
	}
	for {
		parents[int(e.ProcessID)] = int(e.ParentProcessID)
		children[int(e.ParentProcessID)] = append(children[int(e.ParentProcessID)], int(e.ProcessID))
		if err := windows.Process32Next(snap, &e); err != nil {
			break
		}
	}

	var order []processParent
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
		order = append(order, processParent{id, parents[id]}) // post-order => leaves first
	}
	visit(rootPID)
	return order, true
}

type OSKiller struct{}

// OSKiller names processes by creation time, so the responder can tell a
// process that reused a killed PID from the one it killed.
var _ Identifier = OSKiller{}

func (OSKiller) KillPID(pid int) error {
	h, err := openProcess(windows.PROCESS_TERMINATE, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}

// ProcessIdentity returns the creation time of the process that holds pid now
// (FILETIME units). Two processes that held the same PID never share it.
func (OSKiller) ProcessIdentity(pid int) (int64, error) {
	h, err := openProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(h)
	return creationTime(h, pid)
}

// KillIdentity terminates pid only while it is still the process created at
// identity. The open handle holds that process object, so the check and the
// termination apply to the same process even if the PID is reused meanwhile.
func (OSKiller) KillIdentity(pid int, identity int64) error {
	h, err := openProcess(windows.PROCESS_TERMINATE|windows.PROCESS_QUERY_LIMITED_INFORMATION, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	created, err := creationTime(h, pid)
	if err != nil {
		return err
	}
	if identity <= 0 || created != identity {
		return errPIDReused
	}
	return windows.TerminateProcess(h, 1)
}

func protectedPID(pid int) bool { return pid <= 4 || pid == os.Getpid() }

func openProcess(access uint32, pid int) (windows.Handle, error) {
	if protectedPID(pid) {
		return 0, errors.New("refusing response to protected process")
	}
	h, err := windows.OpenProcess(access, false, uint32(pid))
	if err != nil {
		// A PID that no longer exists yields ERROR_INVALID_PARAMETER — the process
		// was already terminated (benign tree-kill race).
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return 0, ErrProcessGone
		}
		return 0, fmt.Errorf("open pid %d: %w", pid, err)
	}
	return h, nil
}

func creationTime(h windows.Handle, pid int) (int64, error) {
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		return 0, fmt.Errorf("process times of pid %d: %w", pid, err)
	}
	return int64(created.HighDateTime)<<32 | int64(created.LowDateTime), nil
}

type OSSuspender struct{}

var (
	ntdll             = windows.NewLazySystemDLL("ntdll.dll")
	procNtSuspendProc = ntdll.NewProc("NtSuspendProcess")
	procNtResumeProc  = ntdll.NewProc("NtResumeProcess")
)
