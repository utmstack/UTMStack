package connectors

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
)

type LoginContext struct {
	IP        string
	UserAgent string
}

type AuthUsecase interface {
	Login(ctx context.Context, input dto.LoginRequest, lc LoginContext) (*dto.LoginResponse, error)
	Me(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
	UpdateMe(ctx context.Context, userID uuid.UUID, input dto.UpdateMeRequest) (*dto.UserResponse, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, input dto.ChangePasswordRequest) error
	Refresh(ctx context.Context, input dto.RefreshRequest, lc LoginContext) (*dto.TokenPair, error)
	Logout(ctx context.Context, input dto.LogoutRequest) error
	ListSessions(ctx context.Context, userID, currentSessionID uuid.UUID) ([]dto.SessionResponse, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeOtherSessions(ctx context.Context, userID, currentSessionID uuid.UUID) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, imageURL string) (*dto.UserResponse, error)
	RequestPasswordReset(ctx context.Context, input dto.ResetPasswordInitRequest) error
	SetPasswordFromChallenge(ctx context.Context, input dto.ResetPasswordFinishRequest) error
}

type ChallengeMailer interface {
	Send(ctx context.Context, purpose domain.ChallengePurpose, to, name, secret string) error
}

type TfaUsecase interface {
	Enroll(ctx context.Context, userID uuid.UUID, input dto.TfaEnrollmentRequest) (*dto.TfaEnrollmentResponse, error)
	VerifyLoginCode(ctx context.Context, input dto.TfaVerifyCodeRequest, lc LoginContext) (*dto.LoginResponse, error)
	Disable(ctx context.Context, userID uuid.UUID, password string) error
	ResetForUser(ctx context.Context, userID uuid.UUID) error

	IssueLoginChallenge(ctx context.Context, u *domain.User) (domain.TfaFactorType, error)
	HasConfirmedFactor(ctx context.Context, userID uuid.UUID) (bool, error)
}

type CreateUserOptions struct {
	Password string
	Invite   bool
}

type UserUsecase interface {
	List(ctx context.Context, q dto.ListUsersQuery) (*dto.UserListResponse, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.UserDetailResponse, error)
	Create(ctx context.Context, input dto.CreateUserRequest, opts CreateUserOptions) (*dto.UserDetailResponse, error)
	Update(ctx context.Context, id uuid.UUID, input dto.UpdateUserRequest) (*dto.UserDetailResponse, error)
	SetStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	AssignRoles(ctx context.Context, id uuid.UUID, input dto.AssignRolesRequest) error
	ResetTfa(ctx context.Context, id uuid.UUID) error
}

type RoleUsecase interface {
	List(ctx context.Context) ([]dto.RoleResponse, error)
	Get(ctx context.Context, id uuid.UUID) (*dto.RoleDetailResponse, error)
	Create(ctx context.Context, input dto.RoleUpsertRequest) (*dto.RoleDetailResponse, error)
	Update(ctx context.Context, id uuid.UUID, input dto.RoleUpsertRequest) (*dto.RoleDetailResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListPermissions(ctx context.Context) ([]dto.PermissionResponse, error)
}

type APIKeyAuthResult struct {
	UserID      uuid.UUID
	Email       string
	Roles       []string
	Permissions []string
	TenantID    uuid.UUID
}

type APIKeyUsecase interface {
	Create(ctx context.Context, userID uuid.UUID, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error)
	Update(ctx context.Context, userID, id uuid.UUID, req dto.APIKeyUpsertRequest) (*dto.APIKeyResponse, error)
	Delete(ctx context.Context, userID, id uuid.UUID) error
	Get(ctx context.Context, userID, id uuid.UUID) (*dto.APIKeyResponse, error)
	List(ctx context.Context, userID uuid.UUID, q dto.ListAPIKeysQuery) (*dto.APIKeyListResponse, error)
	Generate(ctx context.Context, userID, id uuid.UUID) (string, error)
	Authenticate(ctx context.Context, apiKey, clientIP string) (*APIKeyAuthResult, error)
}

type IdentityProviderUsecase interface {
	Create(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error)
	Update(ctx context.Context, req dto.IdentityProviderRequest) (*domain.IdentityProviderConfig, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.IdentityProviderConfig, error)
	List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error)
	ListActive(ctx context.Context) ([]dto.IdentityProviderPublic, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListMappings(ctx context.Context, id uuid.UUID) ([]domain.IdentityProviderGroupMapping, error)
}

type FederationUsecase interface {
	StartURL(ctx context.Context, providerName string) (url string, state string, err error)
	ConsumeSAML(ctx context.Context, providerName string, r *http.Request, requestID string, lc LoginContext) (*dto.TokenPair, error)
	ConsumeOIDC(ctx context.Context, providerName, code, state, wantState string, lc LoginContext) (*dto.TokenPair, error)
	AuthenticateLDAP(ctx context.Context, email, password string, providerID *uuid.UUID, lc LoginContext) (*dto.TokenPair, error)
}
