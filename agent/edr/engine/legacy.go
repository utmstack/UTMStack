package engine

import (
	"os"
	"path/filepath"
	"regexp"
)

// legacyUpdaterState is the state file the retired external signature updater
// kept in its database directory. Its presence identifies the container files
// there as that updater's downloads rather than administrator-provisioned input.
const legacyUpdaterState = "freshclam.dat"

var legacyUpdaterFile = regexp.MustCompile(`(?i)^((main|daily|bytecode)\.(cvd|cld)|(main|daily|bytecode)(-[0-9]+)?\.(cvd|cdiff)\.sign|[a-z0-9_]+-[0-9]+\.cdiff|freshclam\.dat|mirrors\.dat)$`)

// RetireLegacyUpdaterFiles moves the retired updater's downloads out of dir
// into a new sibling "legacy-database-*" directory, so the native engine stops
// trying to compile them with every reload. It does nothing unless the
// updater's state file is present, and it moves files, never deletes them.
func RetireLegacyUpdaterFiles(dir string) (int, error) {
	if _, err := os.Lstat(filepath.Join(dir, legacyUpdaterState)); err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	var names []string
	for _, e := range entries {
		if e.Type().IsRegular() && legacyUpdaterFile.MatchString(e.Name()) {
			names = append(names, e.Name())
		}
	}
	target, err := os.MkdirTemp(filepath.Dir(dir), "legacy-database-")
	if err != nil {
		return 0, err
	}
	moved := 0
	for _, name := range names {
		if err := os.Rename(filepath.Join(dir, name), filepath.Join(target, name)); err != nil {
			return moved, err
		}
		moved++
	}
	return moved, nil
}
