package domain

import "errors"

var (
	ErrNotFound           = errors.New("incident not found")
	ErrAlertAlreadyLinked = errors.New("one or more alerts are already linked to an incident")
)
