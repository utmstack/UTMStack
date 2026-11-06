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
	// ErrVariableUnsecretNeedsValue guards SOAR-F4: unticking "secret" on an
	// existing secret variable must not be a way to read it in clear — the
	// caller has to supply the new (non-secret) value explicitly, rather than
	// have the stored secret silently decrypted and handed back.
	ErrVariableUnsecretNeedsValue = errors.New("unmarking a variable as secret requires supplying its new value")
	ErrIncidentRecordNotFound = errors.New("incident record not found")
	ErrFlowNotFound           = errors.New("flow not found")
	ErrFlowNotStarted         = errors.New("flow is not started; start it first")
	ErrSystemFlowContent      = errors.New("system flow content is read-only")
)
