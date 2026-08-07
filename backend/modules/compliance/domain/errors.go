package domain

import "errors"

var (
	ErrFrameworkNotFound     = errors.New("compliance framework not found")
	ErrControlNotFound       = errors.New("compliance control not found")
	ErrReportNotFound        = errors.New("compliance report config not found")
	ErrScheduleNotFound      = errors.New("compliance report schedule not found")
	ErrSystemOwner           = errors.New("system-owned record cannot be deleted (disable it instead)")
	ErrInvalidCron           = errors.New("invalid cron expression")
	ErrControlExists         = errors.New("a control with this id already exists")
	ErrFrameworkExists       = errors.New("a framework with this key already exists")
	ErrInvalidID             = errors.New("invalid id/key (must be non-empty and contain no path separators)")
	ErrFrameworkLocked       = errors.New("this framework requires an Enterprise license")
	ErrControlLocked         = errors.New("this control requires an Enterprise license")
	ErrInvalidStatus         = errors.New("invalid control status")
	ErrFrameworkNotPossessed = errors.New("this tenant has not enabled this framework")
)
