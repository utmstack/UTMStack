package domain

import "errors"

var (
	ErrNotFound    = errors.New("notification not found")
	ErrInvalidEnum = errors.New("invalid enum value")
)
