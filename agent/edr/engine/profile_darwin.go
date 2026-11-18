//go:build darwin

package engine

import (
	"runtime"

	"golang.org/x/sys/unix"
)

// The macOS sysinfo provider needs CGo and is unavailable in native Go builds.
// Use the kernel's read-only physical-memory query directly, without starting a
// subprocess or importing the provider's C host-information implementation.
func profileHost() (uint64, int, error) {
	ram, err := unix.SysctlUint64("hw.memsize")
	return ram, runtime.NumCPU(), err
}
