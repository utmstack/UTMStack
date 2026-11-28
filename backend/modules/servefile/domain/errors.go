package domain

import "errors"

var (
	// ErrNotFound: no row (or the encrypted blob on disk) for this id.
	ErrNotFound = errors.New("file not found")
	// ErrTenantMismatch: the acting tenant does not own the file.
	ErrTenantMismatch = errors.New("file belongs to another tenant")
)
