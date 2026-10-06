//go:build linux

package ransomware

import (
	"os"
	"strconv"
)

// LiveImager resolves a process image straight from the OS (the live
// snapshot), for PIDs the process table has not recorded yet.
type LiveImager struct{}

// Image reads the /proc/<pid>/exe symlink. Returns an error when the process
// is already gone or unreadable.
func (LiveImager) Image(pid int) (string, error) {
	return os.Readlink("/proc/" + strconv.Itoa(pid) + "/exe")
}
