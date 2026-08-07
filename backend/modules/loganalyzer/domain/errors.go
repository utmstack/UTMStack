package domain

import "errors"

var (
	ErrNotFound        = errors.New("not found")
	ErrIDForbidden     = errors.New("id must be absent on create")
	ErrIDRequired      = errors.New("id is required for update")
	ErrNameRequired    = errors.New("name is required")
	ErrFieldRequired   = errors.New("field is required")
	ErrDatasetRequired = errors.New("dataset is required")
	ErrUnknownDataset  = errors.New("unknown dataset")
	ErrInvalidSQL      = errors.New("invalid SQL")
)

var ErrNoTextSearch = errors.New("this data source has no full-text field; filter by field instead")
