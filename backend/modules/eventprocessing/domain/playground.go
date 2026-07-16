package domain

import "errors"

var (
	ErrPlaygroundBadInput      = errors.New("playground: bad input")
	ErrPlaygroundBusy          = errors.New("playground: busy")
	ErrPlaygroundMisconfigured = errors.New("playground: misconfigured")
	ErrPlaygroundInfra         = errors.New("playground: infra error")
)
