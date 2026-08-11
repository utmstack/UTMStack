package domain

import "errors"

var (
	ErrCorrelationRuleNotFound       = errors.New("correlation rule not found")
	ErrCorrelationRuleSystemOwner    = errors.New("system's rules can't be updated")
	ErrCorrelationRuleNullDefinition = errors.New("correlation rule definition must not be null")
	ErrCorrelationRuleInvalidContent = errors.New("invalid correlation rule")
	ErrIDMustBeAbsent                = errors.New("id must be absent on create")
	ErrIDRequired                    = errors.New("id is required on update")
	ErrDataTypesRequired             = errors.New("rule must have at least one data type")
	ErrPipelineNotFound              = errors.New("pipeline not found")
	ErrPipelineSystemOwner           = errors.New("cannot update system-owned pipeline")
	ErrPipelineInvalidContent        = errors.New("invalid pipeline")
	ErrRegexPatternNotFound          = errors.New("regex pattern not found")
	ErrPlaygroundBadInput            = errors.New("playground: bad input")
	ErrPlaygroundBusy                = errors.New("playground: busy")
	ErrPlaygroundMisconfigured       = errors.New("playground: misconfigured")
	ErrPlaygroundInfra               = errors.New("playground: infra error")
)
