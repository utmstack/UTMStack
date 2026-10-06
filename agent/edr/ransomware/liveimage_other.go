//go:build !windows && !linux

package ransomware

import "errors"

// LiveImager is a no-op on platforms without a per-PID live image source the
// guard currently supports (macOS). The process table is the only source.
type LiveImager struct{}

// Image reports unavailable on this platform.
func (LiveImager) Image(int) (string, error) { return "", errors.New("live image source unavailable on this platform") }
