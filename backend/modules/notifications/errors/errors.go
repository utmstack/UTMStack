package errors

import "errors"

var (
	// ErrNotFound is returned when a notification with the given id does not exist.
	ErrNotFound = errors.New("notification not found")

	// ErrInvalidEnum is returned when a path/query enum value is not one of the
	// allowed values (e.g. a status param outside ACTIVE/HIDDEN/DELETED).
	ErrInvalidEnum = errors.New("invalid enum value")
)
