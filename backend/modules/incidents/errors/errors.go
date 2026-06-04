package errors

import "errors"

var (
	// ErrNotFound is returned when a requested incident entity does not exist.
	ErrNotFound = errors.New("incident not found")

	// ErrAlertAlreadyLinked is returned (HTTP 409) when one or more alertIds
	// are already associated with an existing incident.
	ErrAlertAlreadyLinked = errors.New("one or more alerts are already linked to an incident")
)
