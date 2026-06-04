package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"golang.org/x/crypto/bcrypt"
)

const (
	resetKeyLen = 20
	resetTTL    = 24 * time.Hour
)

const (
	AdminRoleName = "ROLE_ADMIN"
	UserRoleName  = "ROLE_USER"
)

type authUsecase struct {
	userRepo    connectors.UserRepository
	rbacRepo    connectors.RBACRepository
	refreshRepo connectors.RefreshTokenRepository
	signer      *jwt.Signer
	limiter     *ratelimit.LoginLimiter
	refreshTTL  time.Duration
	mailer      connectors.PasswordResetMailer
	tfa         connectors.TfaUsecase
	preAuth     *jwt.PreAuthSigner
	tfaEnabled  bool
}

func NewAuthUsecase(
	userRepo connectors.UserRepository,
	rbacRepo connectors.RBACRepository,
	refreshRepo connectors.RefreshTokenRepository,
	signer *jwt.Signer,
	limiter *ratelimit.LoginLimiter,
	refreshTTL time.Duration,
	mailer connectors.PasswordResetMailer,
	tfa connectors.TfaUsecase,
	preAuth *jwt.PreAuthSigner,
	tfaEnabled bool,
) connectors.AuthUsecase {
	return &authUsecase{
		userRepo:    userRepo,
		rbacRepo:    rbacRepo,
		refreshRepo: refreshRepo,
		signer:      signer,
		limiter:     limiter,
		refreshTTL:  refreshTTL,
		mailer:      mailer,
		tfa:         tfa,
		preAuth:     preAuth,
		tfaEnabled:  tfaEnabled,
	}
}

func (u *authUsecase) Login(ctx context.Context, input dto.LoginRequest, lc connectors.LoginContext) (*dto.LoginResponse, error) {
	if u.limiter != nil && lc.IP != "" && u.limiter.IsBlocked(lc.IP) {
		return nil, domain.ErrLoginBlocked
	}

	user, err := u.userRepo.FindByLoginOrEmail(ctx, input.Login)
	if err != nil {
		return nil, err
	}
	if user == nil {
		u.recordFailure(lc.IP)
		return nil, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		u.recordFailure(lc.IP)
		return nil, domain.ErrInvalidCredentials
	}
	if !user.Activated {
		return nil, domain.ErrUserDeactivated
	}

	if u.tfaEnabled && u.tfa != nil && u.preAuth != nil {
		method := user.TFAMethod
		if method == "" {
			if user.Email == "" {
				return nil, domain.ErrTfaNoEmail
			}
			method = domain.TfaMethodEmail
		}
		if method == domain.TfaMethodEmail {
			if err := u.tfa.IssueLoginEmailChallenge(ctx, user.ID, user.Email, user.FirstName); err != nil {
				return nil, err
			}
		}
		token, _, err := u.preAuth.Sign(user.ID, method)
		if err != nil {
			return nil, err
		}
		if u.limiter != nil && lc.IP != "" {
			u.limiter.Reset(lc.IP)
		}
		return &dto.LoginResponse{
			User:         dto.ToUserResponse(*user),
			TfaRequired:  true,
			TfaMethod:    method,
			PreAuthToken: token,
		}, nil
	}

	pair, err := u.issueTokenPair(ctx, user, lc)
	if err != nil {
		return nil, err
	}

	if u.limiter != nil && lc.IP != "" {
		u.limiter.Reset(lc.IP)
	}

	return &dto.LoginResponse{TokenPair: *pair, User: dto.ToUserResponse(*user)}, nil
}

func (u *authUsecase) recordFailure(ip string) {
	if u.limiter != nil && ip != "" {
		u.limiter.RecordFailure(ip)
	}
}

func (u *authUsecase) Me(ctx context.Context, userID uint64) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	resp := dto.ToUserResponse(*user)
	return &resp, nil
}

func (u *authUsecase) ChangePassword(ctx context.Context, userID uint64, input dto.ChangePasswordRequest) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return domain.ErrCurrentPassword
	}
	if input.NewPassword == input.CurrentPassword {
		return domain.ErrSamePassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hash), false); err != nil {
		return err
	}
	_ = u.refreshRepo.RevokeAllForUser(ctx, user.ID)
	return nil
}

func (u *authUsecase) Refresh(ctx context.Context, input dto.RefreshRequest, lc connectors.LoginContext) (*dto.TokenPair, error) {
	hash := secret.HashSHA256(input.RefreshToken)
	rt, err := u.refreshRepo.FindActiveByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		return nil, domain.ErrInvalidRefresh
	}
	user, err := u.userRepo.FindByID(ctx, rt.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Activated {
		_ = u.refreshRepo.Revoke(ctx, rt.ID)
		return nil, domain.ErrInvalidRefresh
	}

	pair, err := u.issueTokenPair(ctx, user, lc)
	if err != nil {
		return nil, err
	}
	if err := u.refreshRepo.Revoke(ctx, rt.ID); err != nil {
		return nil, err
	}
	return pair, nil
}

func (u *authUsecase) Logout(ctx context.Context, input dto.LogoutRequest) error {
	hash := secret.HashSHA256(input.RefreshToken)
	rt, err := u.refreshRepo.FindActiveByHash(ctx, hash)
	if err != nil {
		return err
	}
	if rt == nil {
		// Already revoked / unknown — treat as success (idempotent).
		return nil
	}
	if err := u.refreshRepo.Revoke(ctx, rt.ID); err != nil {
		return err
	}
	return nil
}

func (u *authUsecase) issueTokenPair(ctx context.Context, user *domain.User, lc connectors.LoginContext) (*dto.TokenPair, error) {
	return IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
}

// IssueTokenPair creates a refresh-token row and signs an access token for the
// given user. Exported so the TFA usecase can mint the post-verify token pair
// without duplicating the refresh-token bookkeeping.
func IssueTokenPair(
	ctx context.Context,
	userRepo connectors.UserRepository,
	refreshRepo connectors.RefreshTokenRepository,
	signer *jwt.Signer,
	refreshTTL time.Duration,
	user *domain.User,
	lc connectors.LoginContext,
) (*dto.TokenPair, error) {
	permissions, err := userRepo.FindPermissionsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	roleRows, err := userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	roles := make([]string, 0, len(roleRows))
	for _, r := range roleRows {
		roles = append(roles, r.Name)
	}
	refresh, err := secret.GenerateOpaque()
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	refreshExp := now.Add(refreshTTL)
	rt := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: secret.HashSHA256(refresh),
		ExpiresAt: refreshExp,
		CreatedAt: now,
		IP:        lc.IP,
		UserAgent: lc.UserAgent,
	}
	if err := refreshRepo.Create(ctx, rt); err != nil {
		return nil, err
	}
	access, accessExp, err := signer.Sign(user.ID, user.Login, user.Email, permissions, roles, rt.ID)
	if err != nil {
		return nil, err
	}
	return &dto.TokenPair{
		AccessToken:      access,
		RefreshToken:     refresh,
		TokenType:        "Bearer",
		ExpiresAt:        accessExp,
		RefreshExpiresAt: refreshExp,
	}, nil
}

func (u *authUsecase) UpdateMe(ctx context.Context, userID uint64, input dto.UpdateMeRequest) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	if input.Email != "" && input.Email != user.Email {
		other, err := u.userRepo.FindByLoginOrEmail(ctx, input.Email)
		if err != nil {
			return nil, err
		}
		if other != nil && other.ID != user.ID {
			return nil, domain.ErrEmailTaken
		}
		user.Email = input.Email
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.LangKey != "" {
		user.LangKey = input.LangKey
	}
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	resp := dto.ToUserResponse(*user)
	return &resp, nil
}

func (u *authUsecase) ListSessions(ctx context.Context, userID, currentSessionID uint64) ([]dto.SessionResponse, error) {
	tokens, err := u.refreshRepo.ListActiveByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.SessionResponse, 0, len(tokens))
	for _, t := range tokens {
		out = append(out, dto.SessionResponse{
			ID:        t.ID,
			IP:        t.IP,
			UserAgent: t.UserAgent,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
			Current:   t.ID == currentSessionID,
		})
	}
	return out, nil
}

func (u *authUsecase) RevokeSession(ctx context.Context, userID, sessionID uint64) error {
	t, err := u.refreshRepo.FindActiveByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if t == nil || t.UserID != userID {
		return domain.ErrSessionNotFound
	}
	if err := u.refreshRepo.Revoke(ctx, sessionID); err != nil {
		return err
	}
	return nil
}

func (u *authUsecase) RevokeOtherSessions(ctx context.Context, userID, currentSessionID uint64) error {
	if err := u.refreshRepo.RevokeAllForUserExcept(ctx, userID, currentSessionID); err != nil {
		return err
	}
	return nil
}

// RequestPasswordReset issues a single-use reset key, persists it on the user
// row, and emails a reset link. Always returns nil for unknown email /
// deactivated accounts so attackers cannot enumerate users by probing.
func (u *authUsecase) RequestPasswordReset(ctx context.Context, input dto.ResetPasswordInitRequest) error {
	user, err := u.userRepo.FindByLoginOrEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	if user == nil || !user.Activated || user.Email == "" {
		return nil
	}

	token, err := secret.GenerateOpaque()
	if err != nil {
		return err
	}
	if len(token) > resetKeyLen {
		token = token[:resetKeyLen]
	}

	now := time.Now().UTC()
	if err := u.userRepo.SetResetKey(ctx, user.ID, token, now); err != nil {
		return err
	}

	if err := u.mailer.SendPasswordReset(ctx, user.Email, user.FirstName, token); err != nil {
		return err
	}
	return nil
}

func (u *authUsecase) FinishPasswordReset(ctx context.Context, input dto.ResetPasswordFinishRequest) error {
	user, err := u.userRepo.FindByResetKey(ctx, input.Key)
	if err != nil {
		return err
	}
	if user == nil || user.ResetDate == nil || time.Since(*user.ResetDate) > resetTTL {
		return domain.ErrInvalidResetKey
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hash), false); err != nil {
		return err
	}
	if err := u.userRepo.ClearResetKey(ctx, user.ID); err != nil {
		return err
	}
	_ = u.refreshRepo.RevokeAllForUser(ctx, user.ID)
	return nil
}

func (u *authUsecase) UpdateAvatar(ctx context.Context, userID uint64, imageURL string) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	user.ImageURL = imageURL
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	resp := dto.ToUserResponse(*user)
	return &resp, nil
}
