package domain

import "errors"

var (
	ErrNotFound         = errors.New("not found")
	ErrIDForbidden      = errors.New("id must be absent on create")
	ErrIDRequired       = errors.New("id is required for update")
	ErrNameRequired     = errors.New("name is required")
	ErrSQLQueryRequired = errors.New("sqlQuery is required")
)
