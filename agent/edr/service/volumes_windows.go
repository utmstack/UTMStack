//go:build windows

package service

import "golang.org/x/sys/windows"

// defaultFixedVolumes enumerates fixed (non-removable, non-network) drives and
// returns them as drive-letter volumes like "C:".
func defaultFixedVolumes() []string {
	var out []string
	mask, err := windows.GetLogicalDrives()
	if err != nil {
		return []string{"C:"}
	}
	for i := 0; i < 26; i++ {
		if mask&(1<<uint(i)) == 0 {
			continue
		}
		letter := string(rune('A' + i))
		root, err := windows.UTF16PtrFromString(letter + `:\`)
		if err != nil {
			continue
		}
		if windows.GetDriveType(root) == windows.DRIVE_FIXED {
			out = append(out, letter+":")
		}
	}
	if len(out) == 0 {
		out = []string{"C:"}
	}
	return out
}
