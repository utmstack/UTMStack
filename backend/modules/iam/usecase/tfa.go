package usecase

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"image/png"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/hotp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
)

const (
	tfaEmailResendCool   = 28 * time.Second
	tfaChallengeTTL      = 10 * time.Minute
	tfaLoginChallengeTTL = 5 * time.Minute

	// Must match totp.Validate's own hardcoded defaults (pquerna/otp/totp) —
	// acceptCode re-derives the matched counter itself instead of calling
	// Validate, so the window it accepts has to stay identical.
	totpPeriod = 30
	totpSkew   = 1
)

type tfaUsecase struct {
	userRepo      connectors.UserRepository
	refreshRepo   connectors.RefreshTokenRepository
	factorRepo    connectors.TfaFactorRepository
	challengeRepo connectors.ChallengeRepository
	mailer        connectors.ChallengeMailer
	signer        *jwt.Signer
	preAuth       *jwt.PreAuthSigner
	refreshTTL    time.Duration
	brand         appconfig_connectors.BrandNameProvider
	limiter       *ratelimit.LoginLimiter
}

func NewTfaUsecase(
	userRepo connectors.UserRepository,
	refreshRepo connectors.RefreshTokenRepository,
	factorRepo connectors.TfaFactorRepository,
	challengeRepo connectors.ChallengeRepository,
	mailer connectors.ChallengeMailer,
	signer *jwt.Signer,
	preAuth *jwt.PreAuthSigner,
	refreshTTL time.Duration,
	brand appconfig_connectors.BrandNameProvider,
	limiter *ratelimit.LoginLimiter,
) connectors.TfaUsecase {
	return &tfaUsecase{
		userRepo:      userRepo,
		refreshRepo:   refreshRepo,
		factorRepo:    factorRepo,
		challengeRepo: challengeRepo,
		mailer:        mailer,
		signer:        signer,
		preAuth:       preAuth,
		refreshTTL:    refreshTTL,
		brand:         brand,
		limiter:       limiter,
	}
}

// tfaLimiterKey scopes the limiter to this user's TFA code attempts,
// separately from Login's IP-keyed budget: the pre-auth token already
// identifies exactly one account, so keying on it (not the caller's IP) stops
// a brute force regardless of how many source addresses it's spread across
// (ACC-F6).
func tfaLimiterKey(userID uuid.UUID) string { return "tfa:" + userID.String() }

func (u *tfaUsecase) HasConfirmedFactor(ctx context.Context, userID uuid.UUID) (bool, error) {
	factors, err := u.factorRepo.ListByUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, f := range factors {
		if f.ConfirmedAt != nil {
			return true, nil
		}
	}
	return false, nil
}

func (u *tfaUsecase) confirmedFactor(ctx context.Context, userID uuid.UUID, t domain.TfaFactorType) (*domain.TfaFactor, error) {
	f, err := u.factorRepo.FindByUserAndType(ctx, userID, t)
	if err != nil {
		return nil, err
	}
	if f == nil || f.ConfirmedAt == nil {
		return nil, nil
	}
	return f, nil
}

func (u *tfaUsecase) IssueLoginChallenge(ctx context.Context, user *domain.User) (domain.TfaFactorType, error) {
	if f, err := u.confirmedFactor(ctx, user.ID, domain.TfaFactorTotp); err != nil {
		return "", err
	} else if f != nil {
		return domain.TfaFactorTotp, nil
	}

	code, err := numericCode()
	if err != nil {
		return "", err
	}
	now := time.Now().UTC()
	existing, err := u.challengeRepo.Get(ctx, domain.ChallengeTfaLogin, user.ID)
	if err != nil {
		return "", err
	}
	if existing != nil && now.Before(existing.CooldownAt) {
		return "", domain.ErrChallengeCooldown
	}

	var factorID *uuid.UUID
	if f, err := u.confirmedFactor(ctx, user.ID, domain.TfaFactorEmail); err != nil {
		return "", err
	} else if f != nil {
		factorID = &f.ID
	}

	c := &domain.UserChallenge{
		UserID:     user.ID,
		TenantID:   user.TenantID,
		Purpose:    domain.ChallengeTfaLogin,
		FactorID:   factorID,
		Secret:     code,
		ExpiresAt:  now.Add(tfaLoginChallengeTTL),
		CooldownAt: now.Add(tfaEmailResendCool),
	}
	if err := u.challengeRepo.Put(ctx, c); err != nil {
		return "", err
	}
	if err := u.mailer.Send(ctx, domain.ChallengeTfaLogin, user.Email, user.Name, code); err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrTfaMailUnavailable, err)
	}
	return domain.TfaFactorEmail, nil
}

func (u *tfaUsecase) VerifyLoginCode(ctx context.Context, input dto.TfaVerifyCodeRequest, lc connectors.LoginContext) (*dto.LoginResponse, error) {
	claims, err := u.preAuth.Verify(input.PreAuthToken)
	if err != nil {
		return nil, domain.ErrTfaInvalidPreAuth
	}
	userID, err := claims.UserID()
	if err != nil {
		return nil, domain.ErrTfaInvalidPreAuth
	}

	key := tfaLimiterKey(userID)
	if u.limiter != nil && u.limiter.IsBlocked(key) {
		return nil, domain.ErrLoginBlocked
	}

	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || user.Status != domain.UserStatusActive {
		return nil, domain.ErrInvalidCredentials
	}

	if err := u.acceptCode(ctx, user.ID, input.Code); err != nil {
		if u.limiter != nil {
			u.limiter.RecordFailure(key)
		}
		return nil, err
	}
	if u.limiter != nil {
		u.limiter.Reset(key)
	}

	pair, err := IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
	if err != nil {
		return nil, err
	}
	_ = u.challengeRepo.Delete(ctx, domain.ChallengeTfaLogin, user.ID)

	return &dto.LoginResponse{TokenPair: *pair, User: dto.ToUserResponse(*user, true)}, nil
}

func (u *tfaUsecase) acceptCode(ctx context.Context, userID uuid.UUID, code string) error {
	factors, err := u.factorRepo.ListByUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, f := range factors {
		if f.ConfirmedAt == nil {
			continue
		}
		switch f.Type {
		case domain.TfaFactorTotp:
			// ACC-F6: totp.Validate is a pure function — the same code stays
			// valid, and keeps re-validating, for its whole ~90s window. Only
			// accepting a counter strictly newer than the last one this factor
			// consumed makes every code single-use.
			if counter, ok := acceptTOTPOnce(code, f.Secret, f.LastUsedAt); ok {
				return u.factorRepo.MarkUsed(ctx, f.ID, time.Unix(int64(counter)*totpPeriod, 0).UTC())
			}
		case domain.TfaFactorRecovery:
			if bcrypt.CompareHashAndPassword([]byte(f.Secret), []byte(code)) == nil {
				return u.factorRepo.Delete(ctx, f.ID)
			}
		}
	}

	c, err := u.challengeRepo.Get(ctx, domain.ChallengeTfaLogin, userID)
	if err != nil {
		return err
	}
	if c == nil {
		return domain.ErrTfaInvalidCode
	}
	if !c.ExpiresAt.After(time.Now().UTC()) {
		return domain.ErrChallengeExpired
	}
	if subtle.ConstantTimeCompare([]byte(c.Secret), []byte(code)) != 1 {
		return domain.ErrTfaInvalidCode
	}
	// The email challenge row itself is the single-use guard here (deleted by
	// VerifyLoginCode once the whole flow succeeds) — no counter to track.
	if c.FactorID == nil {
		return nil
	}
	return u.factorRepo.MarkUsed(ctx, *c.FactorID, time.Now().UTC())
}

// acceptTOTPOnce checks code against the same ±totpSkew window
// totp.Validate(code, secret) would (so legitimate clock drift still works),
// but only accepts a counter strictly greater than the one lastUsedAt maps
// to — rejecting a replay of a code already accepted this window or any
// earlier one (ACC-F6). Returns the matched counter so the caller can persist
// it as the new high-water mark.
func acceptTOTPOnce(code, secret string, lastUsedAt *time.Time) (counter uint64, ok bool) {
	var minCounter uint64
	if lastUsedAt != nil {
		minCounter = uint64(lastUsedAt.Unix())/totpPeriod + 1
	}

	now := uint64(time.Now().UTC().Unix()) / totpPeriod
	found := false
	var best uint64
	for i := -int(totpSkew); i <= int(totpSkew); i++ {
		c := int64(now) + int64(i)
		if c < 0 || uint64(c) < minCounter {
			continue
		}
		candidate := uint64(c)
		valid, err := hotp.ValidateCustom(code, candidate, secret, hotp.ValidateOpts{
			Digits:    otp.DigitsSix,
			Algorithm: otp.AlgorithmSHA1,
		})
		if err != nil || !valid {
			continue
		}
		if !found || candidate > best {
			best, found = candidate, true
		}
	}
	return best, found
}

func (u *tfaUsecase) Enroll(ctx context.Context, userID uuid.UUID, input dto.TfaEnrollmentRequest) (*dto.TfaEnrollmentResponse, error) {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}

	switch input.Stage {
	case domain.TfaStageInit:
		init, err := u.initEnrollment(ctx, user, input.Type)
		if err != nil {
			return nil, err
		}
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Init: init}, nil

	case domain.TfaStageVerify:
		if err := u.verifyEnrollment(ctx, user, input.Type, input.Code); err != nil {
			return nil, err
		}
		verified := true
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Verified: &verified}, nil

	case domain.TfaStageComplete:
		enabled, err := u.HasConfirmedFactor(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Enabled: &enabled}, nil
	}
	return nil, domain.ErrTfaTypeUnsupported
}

func (u *tfaUsecase) initEnrollment(ctx context.Context, user *domain.User, t domain.TfaFactorType) (*dto.TfaInitResponse, error) {
	if existing, err := u.confirmedFactor(ctx, user.ID, t); err != nil {
		return nil, err
	} else if existing != nil {
		return nil, domain.ErrTfaFactorExists
	}

	now := time.Now().UTC()
	resp := &dto.TfaInitResponse{Type: t, ExpiresAt: now.Add(tfaChallengeTTL)}

	switch t {
	case domain.TfaFactorTotp:
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      u.brand.ProductName(ctx),
			AccountName: user.Email,
		})
		if err != nil {
			return nil, err
		}
		factor := &domain.TfaFactor{
			UserID:   user.ID,
			TenantID: user.TenantID,
			Type:     domain.TfaFactorTotp,
			Secret:   key.Secret(),
		}
		if err := u.factorRepo.Create(ctx, factor); err != nil {
			return nil, err
		}
		qr, err := qrDataURL(key.URL())
		if err != nil {
			return nil, err
		}
		resp.FactorID = factor.ID
		resp.OtpAuthURL = key.URL()
		resp.QRDataURL = qr

	case domain.TfaFactorEmail:
		factor := &domain.TfaFactor{
			UserID:   user.ID,
			TenantID: user.TenantID,
			Type:     domain.TfaFactorEmail,
		}
		if err := u.factorRepo.Create(ctx, factor); err != nil {
			return nil, err
		}
		code, err := numericCode()
		if err != nil {
			return nil, err
		}
		c := &domain.UserChallenge{
			UserID:     user.ID,
			TenantID:   user.TenantID,
			Purpose:    domain.ChallengeTfaEnrollment,
			FactorID:   &factor.ID,
			Secret:     code,
			ExpiresAt:  now.Add(tfaChallengeTTL),
			CooldownAt: now.Add(tfaEmailResendCool),
		}
		if err := u.challengeRepo.Put(ctx, c); err != nil {
			return nil, err
		}
		if err := u.mailer.Send(ctx, domain.ChallengeTfaEnrollment, user.Email, user.Name, code); err != nil {
			// The factor was written before the code could be sent. Leaving it
			// behind would litter the account with enrolments that never began,
			// and a later "does this user have 2FA" that forgets to check
			// confirmed_at would lock them out of their own login.
			_ = u.factorRepo.Delete(ctx, factor.ID)
			_ = u.challengeRepo.Delete(ctx, domain.ChallengeTfaEnrollment, user.ID)
			return nil, fmt.Errorf("%w: %v", domain.ErrTfaMailUnavailable, err)
		}
		resp.FactorID = factor.ID
		resp.EmailSent = true

	default:
		return nil, domain.ErrTfaTypeUnsupported
	}
	return resp, nil
}

func (u *tfaUsecase) verifyEnrollment(ctx context.Context, user *domain.User, t domain.TfaFactorType, code string) error {
	factor, err := u.factorRepo.FindByUserAndType(ctx, user.ID, t)
	if err != nil {
		return err
	}
	if factor == nil {
		return domain.ErrTfaFactorNotFound
	}

	switch t {
	case domain.TfaFactorTotp:
		if !totp.Validate(code, factor.Secret) {
			return domain.ErrTfaInvalidCode
		}
	case domain.TfaFactorEmail:
		c, err := u.challengeRepo.Get(ctx, domain.ChallengeTfaEnrollment, user.ID)
		if err != nil {
			return err
		}
		if c == nil {
			return domain.ErrChallengeNotFound
		}
		if !c.ExpiresAt.After(time.Now().UTC()) {
			return domain.ErrChallengeExpired
		}
		if subtle.ConstantTimeCompare([]byte(c.Secret), []byte(code)) != 1 {
			return domain.ErrTfaInvalidCode
		}
	default:
		return domain.ErrTfaTypeUnsupported
	}

	if err := u.factorRepo.Confirm(ctx, factor.ID, time.Now().UTC()); err != nil {
		return err
	}
	return u.challengeRepo.Delete(ctx, domain.ChallengeTfaEnrollment, user.ID)
}

func (u *tfaUsecase) Disable(ctx context.Context, userID uuid.UUID, password string) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return domain.ErrUserNotFound
	}
	if user.PasswordHash == nil {
		return domain.ErrCurrentPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password)); err != nil {
		return domain.ErrCurrentPassword
	}
	return u.ResetForUser(ctx, userID)
}

func (u *tfaUsecase) ResetForUser(ctx context.Context, userID uuid.UUID) error {
	if err := u.factorRepo.DeleteByUser(ctx, userID); err != nil {
		return err
	}
	if err := u.challengeRepo.Delete(ctx, domain.ChallengeTfaEnrollment, userID); err != nil {
		return err
	}
	return u.challengeRepo.Delete(ctx, domain.ChallengeTfaLogin, userID)
}

func numericCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func qrDataURL(otpURL string) (string, error) {
	key, err := otp.NewKeyFromURL(otpURL)
	if err != nil {
		return "", err
	}
	img, err := key.Image(256, 256)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
