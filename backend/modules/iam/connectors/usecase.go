package connectors

import (
	"context"
	"net/http"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type LoginContext struct {
	IP        string
	UserAgent string
}

type AuthUsecase interface {
	Login(ctx context.Context, input dto.LoginRequest, lc LoginContext) (*dto.LoginResponse, error)
	Me(ctx context.Context, userID uint64) (*dto.UserResponse, error)
	UpdateMe(ctx context.Context, userID uint64, input dto.UpdateMeRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID uint64, input dto.ChangePasswordRequest) error
	Refresh(ctx context.Context, input dto.RefreshRequest, lc LoginContext) (*dto.TokenPair, error)
	Logout(ctx context.Context, input dto.LogoutRequest) error
	ListSessions(ctx context.Context, userID, currentSessionID uint64) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID uint64) error
	RevokeOtherSessions(ctx context.Context, userID, currentSessionID uint64) error
	UpdateAvatar(ctx context.Context, userID uint64, imageURL string) (*dto.UserResponse, error)
	RequestPasswordReset(ctx context.Context, input dto.ResetPasswordInitRequest) error
	FinishPasswordReset(ctx context.Context, input dto.ResetPasswordFinishRequest) error
}

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, to, firstName, resetKey string) error
}

type UserInvitationMailer interface {
	SendInvitation(ctx context.Context, to, firstName, resetKey string) error
}

type TfaMailer interface {
	SendTfaCode(ctx context.Context, to, firstName, code string) error
}

type TfaUsecase interface {
	InitEnrollment(ctx context.Context, userID uint64, method string) (*dto.TfaInitResponse, error)
	VerifyEnrollment(ctx context.Context, userID uint64, method, code string) error
	CompleteEnrollment(ctx context.Context, userID uint64, method string) error
	RefreshChallenge(ctx context.Context, userID uint64, method string) (*dto.TfaRefreshResponse, error)
	VerifyLoginCode(ctx context.Context, input dto.TfaVerifyCodeRequest, lc LoginContext) (*dto.LoginResponse, error)
	UnifiedEnrollment(ctx context.Context, userID uint64, input dto.TfaEnrollmentRequest) (*dto.TfaEnrollmentResponse, error)
	IssueLoginEmailChallenge(ctx context.Context, userID uint64, email, firstName string) error
}

type UserUsecase interface {
	List(ctx context.Context, q dto.ListUsersQuery) (*dto.UserListResponse, error)
	Get(ctx context.Context, id uint64) (*dto.UserDetailResponse, error)
	Create(ctx context.Context, actor string, input dto.CreateUserRequest) (*dto.UserDetailResponse, error)
	Update(ctx context.Context, actor string, id uint64, input dto.UpdateUserRequest) (*dto.UserDetailResponse, error)
	Deactivate(ctx context.Context, id uint64) error
	AssignRoles(ctx context.Context, actor string, id uint64, input dto.AssignRolesRequest) error
}

type RoleUsecase interface {
	List(ctx context.Context) ([]dto.RoleResponse, error)
	Get(ctx context.Context, name string) (*dto.RoleDetailResponse, error)
}

type APIKeyAuthResult struct {
	UserID      uint64
	Login       string
	Email       string
	Roles       []string
	Permissions []string
}

type APIKeyUsecase interface {
	Create(ctx context.Context, userID uint64, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error)
	Update(ctx context.Context, userID, id uint64, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error)
	Delete(ctx context.Context, userID, id uint64) error
	Get(ctx context.Context, userID, id uint64) (*dto.APIKeyResponse, error)
	List(ctx context.Context, userID uint64, q dto.ListAPIKeysQuery) (*dto.APIKeyListResponse, error)
	Generate(ctx context.Context, userID, id uint64) (string, error)
	Authenticate(ctx context.Context, apiKey, clientIP string) (*APIKeyAuthResult, error)
}

type IdentityProviderUsecase interface {
	Create(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error)
	Update(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error)
	GetByID(ctx context.Context, id uint64) (*domain.IdentityProviderConfig, error)
	List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error)
	ListActive(ctx context.Context) ([]dto.IdentityProviderPublic, error)
	Delete(ctx context.Context, id uint64) error
}

// SAMLUsecase drives the live SP-initiated SAML2 web SSO flow.
type SAMLUsecase interface {
	// InitiateURL builds the IdP redirect URL (signed AuthnRequest) for the given provider.
	InitiateURL(ctx context.Context, providerName string) (string, error)
	// ConsumeACS validates the SAMLResponse posted to the ACS, maps the NameID to a
	// local (pre-existing) user and returns a freshly issued access token.
	ConsumeACS(ctx context.Context, providerName string, r *http.Request, lc LoginContext) (string, error)
}
