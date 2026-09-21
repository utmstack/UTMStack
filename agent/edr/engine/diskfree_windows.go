//go:build windows

package engine

import "golang.org/x/sys/windows"

// freeDiskBytes returns free bytes available on the filesystem holding path,
// or 0 if it cannot be determined (callers treat 0 as "unknown, skip clamp").
func freeDiskBytes(path string) uint64 {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0
	}
	var freeToCaller, total, totalFree uint64
	if err := windows.GetDiskFreeSpaceEx(p, &freeToCaller, &total, &totalFree); err != nil {
		return 0
	}
	return freeToCaller
}
