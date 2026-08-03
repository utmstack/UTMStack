package domain

import "errors"

var (
	ErrNotFound          = errors.New("tenant not found")
	ErrNameRequired      = errors.New("name is required")
	ErrDomainRequired    = errors.New("domain is required")
	ErrDomainInvalid     = errors.New("domain is not a valid hostname")
	ErrDomainTaken       = errors.New("domain already belongs to another tenant")
	ErrStatusInvalid     = errors.New("status must be ACTIVE, SUSPENDED or TERMINATED")
	ErrAlreadyTerminated = errors.New("tenant is terminated")
)
