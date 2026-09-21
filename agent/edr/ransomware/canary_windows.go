//go:build windows

package ransomware

import "golang.org/x/sys/windows"

// hideFile sets Hidden+System so the decoy is invisible to ordinary users
// (any tamper is therefore attacker-driven, not an accidental user edit).
func hideFile(path string) error {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.SetFileAttributes(p, windows.FILE_ATTRIBUTE_HIDDEN|windows.FILE_ATTRIBUTE_SYSTEM)
}
