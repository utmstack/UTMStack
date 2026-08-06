package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"golang.org/x/crypto/bcrypt"
)

// maxSessionsPerUser bounds what the profile page has to show. The point of that
// list is spotting a session you do not recognise, and a list of two hundred
// hides one rather than revealing it.
const maxSessionsPerUser = 10

const (
	AdminRoleName = "ROLE_ADMIN"
	UserRoleName  = "ROLE_USER"
)

type authUsecase struct {
	userRepo      connectors.UserRepository
	rbacRepo      connectors.RBACRepository
	refreshRepo   connectors.RefreshTokenRepository
	challengeRepo connectors.ChallengeRepository
	signer        *jwt.Signer
	limiter       *ratelimit.LoginLimiter
	refreshTTL    time.Duration
	mailer        connectors.ChallengeMailer
	tfa           connectors.TfaUsecase
	federation    connectors.FederationUsecase
	preAuth       *jwt.PreAuthSigner
	tfaEnabled    bool
}

func NewAuthUsecase(
	userRepo connectors.UserRepository,
	rbacRepo connectors.RBACRepository,
	refreshRepo connectors.RefreshTokenRepository,
	challengeRepo connectors.ChallengeRepository,
	signer *jwt.Signer,
	limiter *ratelimit.LoginLimiter,
	refreshTTL time.Duration,
	mailer connectors.ChallengeMailer,
	tfa connectors.TfaUsecase,
	federation connectors.FederationUsecase,
	preAuth *jwt.PreAuthSigner,
	tfaEnabled bool,
) connectors.AuthUsecase {
	return &authUsecase{
		userRepo:      userRepo,
		rbacRepo:      rbacRepo,
		refreshRepo:   refreshRepo,
		challengeRepo: challengeRepo,
		signer:        signer,
		limiter:       limiter,
		refreshTTL:    refreshTTL,
		mailer:        mailer,
		tfa:           tfa,
		federation:    federation,
		preAuth:       preAuth,
		tfaEnabled:    tfaEnabled,
	}
}

func statusError(s domain.UserStatus) error {
	switch s {
	case domain.UserStatusActive:
		return nil
	case domain.UserStatusPending:
		return domain.ErrUserPending
	case domain.UserStatusSuspended:
		return domain.ErrUserSuspended
	default:
		return domain.ErrUserInactive
	}
}

func (u *authUsecase) Login(ctx context.Context, input dto.LoginRequest, lc connectors.LoginContext) (*dto.LoginResponse, error) {
	if u.limiter != nil && lc.IP != "" && u.limiter.IsBlocked(lc.IP) {
		return nil, domain.ErrLoginBlocked
	}

	user, err := u.userRepo.FindByEmail(ctx, input.Login)
	if err != nil {
		return nil, err
	}
	// A directory account has no local password to compare, and on first sign-in
	// no row here at all, so the bind is tried before giving up.
	if user == nil || user.IdentityProviderID != nil || input.ProviderID != nil {
		if pair, err := u.tryDirectory(ctx, input, lc); err != nil {
			return nil, err
		} else if pair != nil {
			return pair, nil
		}
	}
	if user == nil {
		u.recordFailure(lc.IP)
		return nil, domain.ErrInvalidCredentials
	}
	if user.IdentityProviderID != nil {
		return nil, domain.ErrPasswordLoginUnavailable
	}
	if user.PasswordHash == nil || *user.PasswordHash == "" {
		u.recordFailure(lc.IP)
		return nil, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		u.recordFailure(lc.IP)
		return nil, domain.ErrInvalidCredentials
	}
	if err := statusError(user.Status); err != nil {
		return nil, err
	}

	if u.tfa != nil && u.preAuth != nil {
		required, err := u.tfa.HasConfirmedFactor(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		// A kill switch, not a mandate: the second factor is asked of whoever
		// enrolled one, and development can turn the whole thing off. Forcing it
		// on accounts that have no factor is what leaves a fresh install unable
		// to sign in — the code goes to a mail server nobody has configured yet.
		if required && u.tfaEnabled {
			factorType, err := u.tfa.IssueLoginChallenge(ctx, user)
			if err != nil {
				return nil, err
			}
			token, _, err := u.preAuth.Sign(user.ID, uuid.Nil, string(factorType))
			if err != nil {
				return nil, err
			}
			u.resetLimiter(lc.IP)
			return &dto.LoginResponse{
				User:         dto.ToUserResponse(*user, required),
				TfaRequired:  true,
				TfaType:      factorType,
				PreAuthToken: token,
			}, nil
		}
	}

	pair, err := u.issueTokenPair(ctx, user, lc)
	if err != nil {
		return nil, err
	}
	u.resetLimiter(lc.IP)
	return &dto.LoginResponse{TokenPair: *pair, User: dto.ToUserResponse(*user, false)}, nil
}

// tryDirectory asks the tenant's LDAP providers whether these credentials are
// theirs. It answers nil when none recognised them, which leaves the ordinary
// failure path to say so.
func (u *authUsecase) tryDirectory(
	ctx context.Context, input dto.LoginRequest, lc connectors.LoginContext,
) (*dto.LoginResponse, error) {
	if u.federation == nil {
		return nil, nil
	}
	pair, err := u.federation.AuthenticateLDAP(ctx, input.Login, input.Password, input.ProviderID, lc)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, nil
	}
	user, err := u.userRepo.FindByEmail(ctx, input.Login)
	if err != nil {
		return nil, err
	}
	u.resetLimiter(lc.IP)
	resp := &dto.LoginResponse{TokenPair: *pair}
	if user != nil {
		resp.User = dto.ToUserResponse(*user, false)
	}
	return resp, nil
}

func (u *authUsecase) recordFailure(ip string) {
	if u.limiter != nil && ip != "" {
		u.limiter.RecordFailure(ip)
	}
}

func (u *authUsecase) resetLimiter(ip string) {
	if u.limiter != nil && ip != "" {
		u.limiter.Reset(ip)
	}
}

func (u *authUsecase) Me(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	return u.respond(ctx, user)
}

func (u *authUsecase) respond(ctx context.Context, user *domain.User) (*dto.UserResponse, error) {
	tfa := false
	if u.tfa != nil {
		var err error
		if tfa, err = u.tfa.HasConfirmedFactor(ctx, user.ID); err != nil {
			return nil, err
		}
	}
	resp := dto.ToUserResponse(*user, tfa)
	return &resp, nil
}

func (u *authUsecase) UpdateMe(ctx context.Context, userID uuid.UUID, input dto.UpdateMeRequest) (*dto.UserResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	if input.Email != "" && input.Email != user.Email {
		exists, err := u.userRepo.ExistsByEmail(ctx, input.Email, user.ID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, domain.ErrEmailTaken
		}
		user.Email = input.Email
	}
	if input.Name != "" {
		user.Name = input.Name
	}
	if input.LangKey != "" {
		user.LangKey = input.LangKey
	}
	if err := u.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}
	return u.respond(ctx, user)
}

func (u *authUsecase) UpdateAvatar(ctx context.Context, userID uuid.UUID, imageURL string) (*dto.UserResponse, error) {
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
	return u.respond(ctx, user)
}

func (u *authUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, input dto.ChangePasswordRequest) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrInvalidCredentials
	}
	if user.IdentityProviderID != nil {
		return domain.ErrPasswordLoginUnavailable
	}
	if user.PasswordHash == nil {
		return domain.ErrCurrentPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.CurrentPassword)); err != nil {
		return domain.ErrCurrentPassword
	}
	if input.NewPassword == input.CurrentPassword {
		return domain.ErrSamePassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	_ = u.refreshRepo.RevokeAllForUser(ctx, user.ID)
	return nil
}

func (u *authUsecase) Refresh(ctx context.Context, input dto.RefreshRequest, lc connectors.LoginContext) (*dto.TokenPair, error) {
	rt, err := u.refreshRepo.FindActiveByHash(ctx, secret.HashSHA256(input.RefreshToken))
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
	if user == nil || user.Status != domain.UserStatusActive {
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
	rt, err := u.refreshRepo.FindActiveByHash(ctx, secret.HashSHA256(input.RefreshToken))
	if err != nil {
		return err
	}
	if rt == nil {
		return nil
	}
	return u.refreshRepo.Revoke(ctx, rt.ID)
}

func (u *authUsecase) ListSessions(ctx context.Context, userID, currentSessionID uuid.UUID) ([]dto.SessionResponse, error) {
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

func (u *authUsecase) RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error {
	t, err := u.refreshRepo.FindActiveByID(ctx, sessionID)
	if err != nil {
		return err
	}
	if t == nil || t.UserID != userID {
		return domain.ErrSessionNotFound
	}
	return u.refreshRepo.Revoke(ctx, sessionID)
}

func (u *authUsecase) RevokeOtherSessions(ctx context.Context, userID, currentSessionID uuid.UUID) error {
	if currentSessionID == uuid.Nil {
		return domain.ErrNoActiveSession
	}
	return u.refreshRepo.RevokeAllForUserExcept(ctx, userID, currentSessionID)
}

func (u *authUsecase) RequestPasswordReset(ctx context.Context, input dto.ResetPasswordInitRequest) error {
	user, err := u.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return err
	}
	if user == nil || user.Status != domain.UserStatusActive || user.IdentityProviderID != nil {
		return nil
	}
	// Whether the address exists stays unsaid, so probing cannot enumerate. That
	// the instance has no mail server is not about this user at all, and hiding
	// it only leaves someone waiting for a message that will never arrive.
	if err := issueLinkChallenge(ctx, u.challengeRepo, u.mailer, user, domain.ChallengePasswordReset, resetTTL); err != nil {
		return fmt.Errorf("%w: %v", domain.ErrTfaMailUnavailable, err)
	}
	return nil
}

func (u *authUsecase) SetPasswordFromChallenge(ctx context.Context, input dto.ResetPasswordFinishRequest) error {
	c, err := resolveLinkChallenge(ctx, u.challengeRepo, input.Key)
	if err != nil {
		return err
	}
	user, err := u.userRepo.FindByID(ctx, c.UserID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := u.userRepo.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		return err
	}
	if c.Purpose == domain.ChallengeActivation || user.Status == domain.UserStatusPending {
		if err := u.userRepo.UpdateStatus(ctx, user.ID, domain.UserStatusActive); err != nil {
			return err
		}
	}
	if err := u.challengeRepo.Delete(ctx, c.Purpose, user.ID); err != nil {
		return err
	}
	_ = u.refreshRepo.RevokeAllForUser(ctx, user.ID)
	return nil
}

func (u *authUsecase) issueTokenPair(ctx context.Context, user *domain.User, lc connectors.LoginContext) (*dto.TokenPair, error) {
	return IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
}

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
		TenantID:  user.TenantID,
		TokenHash: secret.HashSHA256(refresh),
		ExpiresAt: refreshExp,
		CreatedAt: now,
		IP:        lc.IP,
		UserAgent: lc.UserAgent,
	}
	if err := refreshRepo.Create(ctx, rt); err != nil {
		return nil, err
	}

	if _, err := refreshRepo.RevokeOldestBeyond(ctx, user.ID, maxSessionsPerUser); err != nil {
		_ = catcher.Error("could not trim old sessions", err, map[string]any{"user": user.Email})
	}
	tenantID := user.TenantID.String()
	access, accessExp, err := signer.Sign(user.ID, user.Email, permissions, roles, rt.ID, tenantID,
		authz.IsPlatformIdentity(tenantID, roles))
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
