//go:build !windows && !linux

package netblock

// NewOSBlocker cannot enforce on this platform; callers must report unavailable
// protection. Unit tests inject a fake rather than treating a no-op as success.
func NewOSBlocker(sublayerName string) (Blocker, error) { return nil, ErrUnsupported }
