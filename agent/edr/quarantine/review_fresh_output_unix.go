//go:build linux || darwin

package quarantine

import (
	"golang.org/x/sys/unix"
	"os"
)

func openFreshOutput(stage *restoreDirectory) (*os.File, error) {
	return openFreshWritableFile(stage, "snapshot")
}

func openFreshWritableFile(d *restoreDirectory, name string) (*os.File, error) {
	return d.open(name, unix.O_RDWR, 0)
}

func pinFreshDatabaseFile(d *restoreDirectory, name string) (*os.File, error) {
	return d.open(name, unix.O_RDONLY, 0)
}
