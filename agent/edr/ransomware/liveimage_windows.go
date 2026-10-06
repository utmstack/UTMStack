package ransomware

import "golang.org/x/sys/windows"

// LiveImager resolves a process image straight from the OS (the live
// snapshot), for PIDs the process table has not recorded yet.
type LiveImager struct{}

// Image opens the process and asks the kernel for its full image path.
// Returns an error when the process is already gone or unreadable.
func (LiveImager) Image(pid int) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)
	size := uint32(windows.MAX_PATH)
	buf := make([]uint16, size)
	for {
		err = windows.QueryFullProcessImageName(h, 0, &buf[0], &size)
		if err == windows.ERROR_MORE_DATA {
			buf = make([]uint16, size)
			continue
		}
		if err != nil {
			return "", err
		}
		return windows.UTF16ToString(buf[:size]), nil
	}
}
