//go:build windows

package responder

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

// LiveProcess reads a running process's full image path and parent PID from
// the operating system. It serves callers that act before the WMI-fed process
// table has recorded a process that started moments ago.
func LiveProcess(pid int) (string, int, error) {
	h, err := openProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, pid)
	if err != nil {
		return "", 0, err
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	n := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil {
		return "", 0, err
	}
	image := windows.UTF16ToString(buf[:n])
	if image == "" {
		return "", 0, errors.New("empty process image path")
	}
	return image, parentPID(pid), nil
}

// parentPID returns pid's parent from a process snapshot, or 0 if unknown.
func parentPID(pid int) int {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(snap)
	var e windows.ProcessEntry32
	e.Size = uint32(unsafe.Sizeof(e))
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		if int(e.ProcessID) == pid {
			return int(e.ParentProcessID)
		}
	}
	return 0
}
