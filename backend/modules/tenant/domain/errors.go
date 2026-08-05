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
	ErrSupportInvalid    = errors.New("support access must be NONE, READ or FULL")
	ErrNotOwnTenant      = errors.New("only an administrator of this tenant may change its support access")
	ErrLimitNegative     = errors.New("a limit cannot be negative; use 0 for no limit")
	ErrDefaultTenant     = errors.New("the default tenant holds the platform plane and cannot be changed this way")
)
