//go:build !linux

package netblock

func NewHostIsolator() (*HostIsolator, error) { return nil, ErrUnsupported }
