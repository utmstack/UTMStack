//go:build !windows

package netblock

import "net/netip"

type noopBlocker struct{}

// NewOSBlocker returns a no-op Blocker on non-Windows hosts so the package
// builds and unit-tests everywhere. The real WFP implementation lives in
// wfp_windows.go.
func NewOSBlocker(sublayerName string) (Blocker, error) { return &noopBlocker{}, nil }

func (noopBlocker) AddIP(netip.Addr, string) error { return nil }
func (noopBlocker) RemoveIP(netip.Addr) error      { return nil }
func (noopBlocker) Reset() error                   { return nil }
func (noopBlocker) Count() int                     { return 0 }
