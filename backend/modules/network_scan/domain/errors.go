package domain

import "errors"

var (
	ErrNotFound          = errors.New("network_scan: not found")
	ErrInvalidInput      = errors.New("network_scan: invalid input")
	ErrUnknownProperty   = errors.New("network_scan: unknown property")
	ErrProbeUnavailable  = errors.New("network_scan: probe unavailable")
	ErrProbeNotConfigured = errors.New("network_scan: probe url not configured")
)
