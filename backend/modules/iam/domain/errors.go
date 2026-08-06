package domain

import "errors"

// Authentication & session errors.
var (
	ErrInvalidCredentials       = errors.New("invalid credentials")
	ErrUserPending              = errors.New("account has not been activated")
	ErrUserInactive             = errors.New("account is inactive")
	ErrUserSuspended            = errors.New("account is suspended")
	ErrLoginBlocked             = errors.New("login blocked due to repeated failures")
	ErrPasswordLoginUnavailable = errors.New("account authenticates through an identity provider")
	ErrCurrentPassword          = errors.New("current password is incorrect")
	ErrSamePassword             = errors.New("new password must differ from current")
	ErrInvalidRefresh           = errors.New("invalid or expired refresh token")
	ErrSessionNotFound          = errors.New("session not found")
	ErrNoActiveSession          = errors.New("no active session to preserve")
)

// Challenge errors. Activation, password reset and 2FA share UserChallenge, so
// they share these.
var (
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrChallengeExpired  = errors.New("challenge has expired")
	ErrChallengeCooldown = errors.New("too many requests; wait before retrying")
	ErrChallengeCodeUsed = errors.New("code has already been used")
)

// Two-factor authentication errors.
var (
	ErrTfaRequired          = errors.New("tfa verification required")
	ErrTfaInvalidCode       = errors.New("invalid tfa code")
	ErrTfaInvalidPreAuth    = errors.New("invalid or expired pre-auth token")
	ErrTfaFactorNotFound    = errors.New("tfa factor not found")
	ErrTfaFactorExists      = errors.New("a tfa factor of this type is already enrolled")
	ErrTfaFactorUnconfirmed = errors.New("tfa factor has not been confirmed")
	ErrTfaFactorMismatch    = errors.New("tfa factor does not match the pre-auth token")
	ErrTfaTypeUnsupported   = errors.New("unsupported tfa factor type")
	ErrTfaNoEmail           = errors.New("tfa is required but the user has no email on file")

	// ErrTfaMailUnavailable is what an install with no mail server gets. It has
	// to be distinguishable from a wrong password: the administrator seeing it
	// needs to know the fix is in the configuration, not in their fingers.
	ErrTfaMailUnavailable = errors.New("the second factor is sent by email, and no mail server is configured")
	ErrTfaDisabled        = errors.New("tfa is disabled for this tenant")
)

// User management errors.
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrEmailTaken        = errors.New("email is already in use")
	ErrInvalidRoleSubset = errors.New("one or more roles do not exist")
)

// Role & permission errors.
var (
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleNameTaken      = errors.New("role name is already in use")
	ErrRoleImmutable      = errors.New("system roles cannot be modified")
	ErrPermissionNotFound = errors.New("permission not found")
)

// API key errors.
var (
	ErrAPIKeyNameTaken = errors.New("api key name already in use")
	ErrAPIKeyNotFound  = errors.New("api key not found")
	ErrAPIKeyInvalid   = errors.New("invalid api key")
)

// Identity provider errors.
var (
	ErrIDPNotFound          = errors.New("identity provider not found")
	ErrIDPIDForbidden       = errors.New("id must be absent on create")
	ErrIDPIDRequired        = errors.New("id is required for update")
	ErrIDPInvalidInput      = errors.New("name, provider type and the settings that type requires are missing or invalid")
	ErrIDPTypeUnsupported   = errors.New("unsupported identity provider type")
	ErrIDPKeyRequired       = errors.New("sp private key is required on create")
	ErrIDPSettingsInvalid   = errors.New("the settings this provider type requires are missing or invalid")
	ErrIDPStateInvalid      = errors.New("the sign-in did not come back with the state it left with")
	ErrFederatedUnknownUser = errors.New("no local account matches this identity")
	ErrFederatedNoRoles     = errors.New("this identity maps to no role, and the provider has no default")
	ErrSSONotEntitled       = errors.New("single sign-on requires a paid plan")
)
