package correrrors

import "errors"

var (
	// ErrRegexPatternNotFound is returned when a regex pattern ID does not exist.
	ErrRegexPatternNotFound = errors.New("regex pattern not found")

	// ErrRegexPatternSystemOwner is returned when an attempt is made to mutate
	// a system-owned regex pattern.
	ErrRegexPatternSystemOwner = errors.New("system regex pattern cannot be modified or deleted")

	// ErrTenantConfigNotFound is returned when a tenant config ID does not exist.
	ErrTenantConfigNotFound = errors.New("tenant config not found")

	// ErrDataTypeNotFound is returned when a data type ID does not exist.
	ErrDataTypeNotFound = errors.New("data type not found")

	// ErrDataTypeSystemOwner is returned when an attempt is made to mutate
	// a system-owned data type.
	ErrDataTypeSystemOwner = errors.New("system data type cannot be modified or deleted")

	// ErrCorrelationRuleNotFound is returned when a correlation rule ID does not exist.
	ErrCorrelationRuleNotFound = errors.New("correlation rule not found")

	// ErrCorrelationRuleSystemOwner is returned when an attempt is made to mutate
	// a system-owned correlation rule.
	ErrCorrelationRuleSystemOwner = errors.New("system's rules can't be updated")

	// ErrCorrelationRuleNullDefinition is returned when a rule update is sent
	// with a null/empty definition field.
	ErrCorrelationRuleNullDefinition = errors.New("correlation rule definition must not be null")

	// ErrIDMustBeAbsent is returned on POST when the caller supplies an id field.
	ErrIDMustBeAbsent = errors.New("id must be absent on create")

	// ErrIDRequired is returned on PUT when id is null/zero.
	ErrIDRequired = errors.New("id is required on update")

	// ErrDataTypesRequired is returned when a correlation rule create or update
	// is attempted with an empty dataTypes list.
	ErrDataTypesRequired = errors.New("rule must have at least one data type")
)
