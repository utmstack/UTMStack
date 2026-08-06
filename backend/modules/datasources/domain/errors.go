package domain

import "errors"

var (
	ErrNotFound           = errors.New("datasource: not found")
	ErrInvalidInput       = errors.New("datasource: invalid input")
	ErrUnknownProperty    = errors.New("datasource: unknown property")
	ErrProbeUnavailable   = errors.New("datasource: probe unavailable")
	ErrProbeNotConfigured = errors.New("datasource: probe url not configured")
)
