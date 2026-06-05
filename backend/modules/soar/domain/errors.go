package domain

import "errors"

var (
	// ErrRuleNotFound is returned when a rule ID does not exist in the database.
	ErrRuleNotFound = errors.New("alert response rule not found")

	// ErrTemplateNotFound is returned when a template ID does not exist.
	ErrTemplateNotFound = errors.New("alert response action template not found")

	// ErrRuleNameTaken is returned when a rule with the same name already exists.
	ErrRuleNameTaken = errors.New("alert response rule name already in use")

	// ErrIDMustBeAbsent is returned on POST when a caller supplies an id field.
	ErrIDMustBeAbsent = errors.New("id must be absent on create")

	// ErrIDRequired is returned on PUT when id is null/zero.
	ErrIDRequired = errors.New("id is required on update")

	// ErrNotImplemented is returned when a rule filter operator is not supported.
	ErrNotImplemented = errors.New("operator not implemented")

	// ErrVariableNotFound is returned when an incident variable ID does not exist.
	ErrVariableNotFound = errors.New("incident variable not found")

	// ErrIncidentRecordNotFound is returned when an incident action/command/job
	// ID does not exist in the database.
	ErrIncidentRecordNotFound = errors.New("incident record not found")
)
