package domain

import "errors"

var (
	ErrRegexPatternNotFound          = errors.New("regex pattern not found")
	ErrRegexPatternSystemOwner       = errors.New("system regex pattern cannot be modified or deleted")
	ErrTenantConfigNotFound          = errors.New("tenant config not found")
	ErrCorrelationRuleNotFound       = errors.New("correlation rule not found")
	ErrCorrelationRuleSystemOwner    = errors.New("system's rules can't be updated")
	ErrCorrelationRuleNullDefinition = errors.New("correlation rule definition must not be null")
	ErrCorrelationRuleInvalidContent = errors.New("invalid correlation rule")
	ErrIDMustBeAbsent                = errors.New("id must be absent on create")
	ErrIDRequired                    = errors.New("id is required on update")
	ErrDataTypesRequired             = errors.New("rule must have at least one data type")
	ErrFilterIDExists                = errors.New("a new filter cannot already have an ID")
	ErrFilterIDNull                  = errors.New("filter ID must be set for update")
	ErrFilterNotFound                = errors.New("filter not found")
	ErrFilterSystemOwner             = errors.New("cannot update system-owned filter")
	ErrFilterInvalidContent          = errors.New("invalid filter")
)
