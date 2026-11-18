//go:build !windows

package fanotify

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func duplicateFixtureFile(f *os.File) (*os.File, error) {
	root := "/dev/fd"
	if runtime.GOOS == "linux" {
		root = "/proc/self/fd"
	}
	return os.Open(filepath.Join(root, fmt.Sprint(f.Fd())))
}
