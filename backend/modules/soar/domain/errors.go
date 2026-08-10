package domain

import "errors"

var (
	ErrRuleNameTaken          = errors.New("alert response rule name already in use")
	ErrSystemRuleReadOnly     = errors.New("system alert response rule is read-only")
	ErrIDMustBeAbsent         = errors.New("id must be absent on create")
	ErrIDRequired             = errors.New("id is required on update")
	ErrNotImplemented         = errors.New("operator not implemented")
	ErrVariableNotFound       = errors.New("incident variable not found")
	ErrVariableValueRequired  = errors.New("a variable must have a value")
	ErrIncidentRecordNotFound = errors.New("incident record not found")
	ErrFlowNotFound           = errors.New("flow not found")
	ErrSystemFlowContent      = errors.New("system flow content is read-only")
)
