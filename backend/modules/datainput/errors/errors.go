package errors

import "errors"

var (
	// ErrIDExists is returned on POST when the caller supplies an id field.
	ErrIDExists = errors.New("a new data input status cannot already have an ID")

	// ErrIDNull is returned on PUT when the id field is absent.
	ErrIDNull = errors.New("data input status ID must be set for update")

	// ErrDataInputStatusNotFound is returned when a record with the given id does not exist.
	ErrDataInputStatusNotFound = errors.New("data input status not found")
)
