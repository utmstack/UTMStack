package domain

import "errors"

var (
	ErrFrameworkNotFound = errors.New("compliance framework not found")
	ErrControlNotFound   = errors.New("compliance control not found")
	ErrReportNotFound    = errors.New("compliance report not found")
	ErrScheduleNotFound  = errors.New("compliance report schedule not found")
	ErrSystemOwner       = errors.New("the shipped catalogue is read-only — create your own instead of changing it")
	ErrInvalidCron       = errors.New("invalid cron expression")
	ErrControlExists     = errors.New("a control with this id already exists")
	ErrFrameworkExists   = errors.New("a framework with this key already exists")
	ErrInvalidID         = errors.New("invalid id/key (must be non-empty and contain no path separators)")
	ErrFrameworkLocked   = errors.New("this framework requires an Enterprise license")
	ErrControlLocked     = errors.New("this control requires an Enterprise license")
	ErrInvalidStatus     = errors.New("invalid control status")
	ErrReportConflict    = errors.New("this report was modified by another run — reload and try again")
)
