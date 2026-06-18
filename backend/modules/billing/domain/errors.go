package domain

import "errors"

// ErrInvalidLicense is returned when an uploaded LICENSE envelope fails to
// decrypt/verify against this instance, cannot be parsed, or is already expired.
var ErrInvalidLicense = errors.New("invalid license")
