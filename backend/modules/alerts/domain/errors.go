package domain

import "errors"

var (
	ErrAlertTagNotFound     = errors.New("alert tag not found")
	ErrAlertTagRuleNotFound = errors.New("alert tag rule not found")
	ErrTagNameTaken         = errors.New("tag name already in use")
	ErrRuleNameTaken        = errors.New("rule name already in use")
	ErrInvalidAlertStatus   = errors.New("invalid alert status")
	ErrMissingAlertID       = errors.New("missing alert id")
)
