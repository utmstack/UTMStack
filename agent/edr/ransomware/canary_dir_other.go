//go:build !linux && !darwin && !windows

package ransomware

import (
	"errors"
	"os"
)

type canaryDir struct{}

func openCanaryDir(string) (*canaryDir, error) {
	return nil, errors.New("safe canary planting unavailable on this platform")
}
func (*canaryDir) Close() error { return nil }
func (*canaryDir) open(string, bool) (*os.File, error) {
	return nil, errors.New("safe canary planting unavailable on this platform")
}
