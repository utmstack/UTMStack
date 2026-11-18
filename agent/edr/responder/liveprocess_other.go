//go:build !windows && !linux

package responder

import "errors"

// LiveProcess is only implemented on Windows, the supported endpoint platform.
func LiveProcess(pid int) (string, int, error) {
	return "", 0, errors.New("live process lookup is not supported on this platform")
}
