package domain

import "errors"

var (
	ErrNotFound           = errors.New("incident not found")
	ErrAlertAlreadyLinked = errors.New("one or more alerts are already linked to an incident")
	ErrNameTaken          = errors.New("an incident with this name already exists")
	ErrInvalidStatus      = errors.New("unknown incident status")
)
