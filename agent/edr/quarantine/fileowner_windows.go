//go:build windows

package quarantine

// fileOwner is a no-op on Windows: POSIX-style numeric ownership does not
// apply, so callers treat (0, 0, nil) as "unknown" and skip re-applying it.
func fileOwner(path string) (int, int, error) {
	return 0, 0, nil
}
