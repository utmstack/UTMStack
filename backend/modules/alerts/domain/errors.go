package domain

import "errors"

var (
	ErrAlertNotFound        = errors.New("alert not found")
	ErrAlertTagNotFound     = errors.New("alert tag not found")
	ErrAlertTagSystemOwned  = errors.New("system-owned alert tag cannot be modified or deleted")
	ErrAlertTagRuleNotFound = errors.New("alert tag rule not found")
	ErrTagNameTaken         = errors.New("tag name already in use")
	ErrRuleNameTaken        = errors.New("rule name already in use")
	ErrInvalidAlertStatus   = errors.New("invalid alert status")
	ErrMissingAlertID       = errors.New("missing alert id")

	// ErrTooManyAlerts is a refusal, not a failure: an action over more alerts
	// than one request may rewrite is rejected whole rather than applied to
	// part of the selection and reported as done.
	ErrTooManyAlerts = errors.New("this action covers more alerts than one request may change")
)
