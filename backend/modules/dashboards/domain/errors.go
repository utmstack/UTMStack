package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrIDForbidden         = errors.New("id must be absent on create")
	ErrIDRequired          = errors.New("id is required for update")
	ErrNameRequired        = errors.New("name is required")
	ErrSpecRequired        = errors.New("spec is required")
	ErrDashboardIDRequired = errors.New("dashboardId is required")
	ErrSystemOwned         = errors.New("system-owned and cannot be modified")
	ErrInvalidSpec         = errors.New("invalid spec")
)
