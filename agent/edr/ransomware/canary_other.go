//go:build !windows

package ransomware

// hideFile is a no-op off Windows (canary logic is unit-tested on the host).
func hideFile(path string) error { return nil }
