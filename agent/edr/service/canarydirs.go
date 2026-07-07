package service

import (
	"os"
	"path/filepath"
)

// defaultCanaryDirs returns the baseline directories seeded with canaries:
// each watched volume root plus common user-data folders under the profile.
// Extra dirs from config are appended by the caller. defaultFixedVolumes() is
// the same package-local helper the watcher uses (volumes_windows.go / _other.go).
func defaultCanaryDirs() []string {
	var dirs []string
	// defaultFixedVolumes() returns bare drive specs ("C:") for the USN watcher.
	// Anchor each to the drive ROOT before planting: filepath.Join("C:", name) is
	// drive-RELATIVE ("C:name") on Windows, which would plant the decoy in the
	// service CWD and produce a stored path that never matches the ETW feed. With
	// a trailing separator, filepath.Join("C:\\", name) is the absolute root path.
	for _, v := range defaultFixedVolumes() {
		if v == "" {
			continue
		}
		if !os.IsPathSeparator(v[len(v)-1]) {
			v += string(os.PathSeparator)
		}
		dirs = append(dirs, v)
	}
	if up := os.Getenv("USERPROFILE"); up != "" {
		for _, sub := range []string{"Desktop", "Documents", "Downloads", "Pictures"} {
			dirs = append(dirs, filepath.Join(up, sub))
		}
	}
	return dirs
}
