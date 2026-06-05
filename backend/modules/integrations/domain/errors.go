package domain

import "errors"

var (
	ErrTenantNotFound      = errors.New("tenant not found")
	ErrTenantExists        = errors.New("tenant already exists")
	ErrRequiredConfigEmpty = errors.New("required config field is empty")

	ErrModuleNotFound    = errors.New("module not found")
	ErrInvalidActivation = errors.New("invalid activation request")
	ErrSystemModule      = errors.New("system integrations cannot be modified or deleted")
	ErrNotConfigurable   = errors.New("integration is not configurable (no schema / has no tenants)")
)
