package domain

import "errors"

// Authentication & session errors.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserDeactivated    = errors.New("user is deactivated")
	ErrLoginBlocked       = errors.New("login blocked due to repeated failures")
	ErrCurrentPassword    = errors.New("current password is incorrect")
	ErrSamePassword       = errors.New("new password must differ from current")
	ErrInvalidRefresh     = errors.New("invalid or expired refresh token")
	ErrEmailTaken         = errors.New("email is already in use")
	ErrSessionNotFound    = errors.New("session not found")
	ErrNoActiveSession    = errors.New("no active session to preserve")
	ErrInvalidResetKey    = errors.New("invalid or expired reset key")
)

// Two-factor authentication errors.
var (
	ErrTfaRequired          = errors.New("tfa verification required")
	ErrTfaChallengeNotFound = errors.New("tfa challenge not found or expired")
	ErrTfaInvalidCode       = errors.New("invalid or expired tfa code")
	ErrTfaCooldown          = errors.New("too many tfa requests; wait before retrying")
	ErrTfaMethodUnsupported = errors.New("unsupported tfa method")
	ErrTfaNotVerified       = errors.New("tfa challenge has not been verified yet")
	ErrTfaInvalidPreAuth    = errors.New("invalid or expired pre-auth token")
	ErrTfaAlreadyEnabled    = errors.New("tfa is already enabled")
	ErrTfaNotEnabled        = errors.New("tfa is not enabled for this user")
	ErrTfaMethodMismatch    = errors.New("tfa method does not match pre-auth token")
	ErrTfaDisabled          = errors.New("tfa is globally disabled")
	ErrTfaNoEmail           = errors.New("tfa is required but the user has no email on file")
)

// User management errors.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrLoginOrEmailTaken = errors.New("login or email already in use")
	ErrInvalidRoleSubset = errors.New("one or more role names do not exist")
)

// Role errors.
var (
	ErrRoleNotFound = errors.New("role not found")
)

// API key errors.
var (
	ErrAPIKeyNameTaken = errors.New("api key name already in use")
	ErrAPIKeyNotFound  = errors.New("api key not found")
	ErrAPIKeyInvalid   = errors.New("invalid api key")
)
