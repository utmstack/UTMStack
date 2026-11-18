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
	dirs = append(dirs, optionalProfileCanaryDirs(os.Getenv("USERPROFILE"), os.Lstat)...)
	return dirs
}

// SYSTEM profiles on server installations commonly lack these optional user
// folders. Absence is not a failed configured canary target. Existing unsafe
// objects and access/inspection errors still reach the pinned directory
// validator so they degrade coverage rather than being silently excluded.
// This is the service account's profile only, not coverage of all user profiles.
func optionalProfileCanaryDirs(profile string, lstat func(string) (os.FileInfo, error)) []string {
	if profile == "" {
		return nil
	}
	var dirs []string
	for _, sub := range []string{"Desktop", "Documents", "Downloads", "Pictures"} {
		path := filepath.Join(profile, sub)
		if _, err := lstat(path); !os.IsNotExist(err) {
			dirs = append(dirs, path)
		}
	}
	return dirs
}
