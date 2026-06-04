package domain

import "errors"

var (
	ErrModuleNotFound       = errors.New("module not found")
	ErrGroupNotFound        = errors.New("module group not found")
	ErrConfigNotFound       = errors.New("module group configuration not found")
	ErrModuleKindNotPorted  = errors.New("module kind not yet ported")
	ErrInvalidActivation    = errors.New("invalid activation request")
	ErrRequiredConfigEmpty  = errors.New("required configuration value is empty")
	ErrModuleConfigConflict = errors.New("conflicting module group configuration")
)
