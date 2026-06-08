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

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
)

const (
	tfaEmailResendCool   = 28 * time.Second
	tfaChallengeTTL      = 10 * time.Minute
	tfaLoginChallengeTTL = 5 * time.Minute
)

type tfaUsecase struct {
	userRepo    connectors.UserRepository
	refreshRepo connectors.RefreshTokenRepository
	rbacRepo    connectors.RBACRepository
	stateRepo   connectors.TfaStateRepository
	mailer      connectors.TfaMailer
	signer      *jwt.Signer
	preAuth     *jwt.PreAuthSigner
	refreshTTL  time.Duration
	enabled     bool
	brand       appconfig_connectors.BrandNameProvider
}

func NewTfaUsecase(
	userRepo connectors.UserRepository,
	refreshRepo connectors.RefreshTokenRepository,
	rbacRepo connectors.RBACRepository,
	stateRepo connectors.TfaStateRepository,
	mailer connectors.TfaMailer,
	signer *jwt.Signer,
	preAuth *jwt.PreAuthSigner,
	refreshTTL time.Duration,
	enabled bool,
	brand appconfig_connectors.BrandNameProvider,
) connectors.TfaUsecase {
	return &tfaUsecase{
		userRepo:    userRepo,
		refreshRepo: refreshRepo,
		rbacRepo:    rbacRepo,
		stateRepo:   stateRepo,
		mailer:      mailer,
		signer:      signer,
		preAuth:     preAuth,
		refreshTTL:  refreshTTL,
		enabled:     enabled,
		brand:       brand,
	}
}

func (u *tfaUsecase) InitEnrollment(ctx context.Context, userID uint64, method string) (*dto.TfaInitResponse, error) {
	if !u.enabled {
		return nil, domain.ErrTfaDisabled
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}
	if user.TFAMethod != "" {
		return nil, domain.ErrTfaAlreadyEnabled
	}

	switch method {
	case domain.TfaMethodTotp:
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      u.brand.ProductName(ctx),
			AccountName: tfaAccountName(user),
		})
		if err != nil {
			return nil, err
		}
		state := &domain.TfaSetupState{
			UserID:    user.ID,
			Purpose:   domain.TfaPurposeEnrollment,
			Method:    domain.TfaMethodTotp,
			Secret:    key.Secret(),
			ExpiresAt: time.Now().Add(tfaChallengeTTL),
		}
		if err := u.stateRepo.Put(ctx, state); err != nil {
			return nil, err
		}
		dataURL, err := renderQRDataURL(key.URL())
		if err != nil {
			return nil, err
		}
		return &dto.TfaInitResponse{
			Method:     domain.TfaMethodTotp,
			QRDataURL:  dataURL,
			OtpAuthURL: key.URL(),
			ExpiresAt:  state.ExpiresAt,
		}, nil

	case domain.TfaMethodEmail:
		if user.Email == "" {
			return nil, domain.ErrTfaMethodUnsupported
		}
		code, err := newSixDigitCode()
		if err != nil {
			return nil, err
		}
		state := &domain.TfaSetupState{
			UserID:     user.ID,
			Purpose:    domain.TfaPurposeEnrollment,
			Method:     domain.TfaMethodEmail,
			Secret:     code,
			ExpiresAt:  time.Now().Add(tfaChallengeTTL),
			CooldownAt: time.Now().Add(tfaEmailResendCool),
		}
		if err := u.stateRepo.Put(ctx, state); err != nil {
			return nil, err
		}
		if err := u.mailer.SendTfaCode(ctx, user.Email, user.FirstName, code); err != nil {
			return nil, err
		}
		return &dto.TfaInitResponse{
			Method:    domain.TfaMethodEmail,
			EmailSent: true,
			ExpiresAt: state.ExpiresAt,
		}, nil

	default:
		return nil, domain.ErrTfaMethodUnsupported
	}
}

func (u *tfaUsecase) VerifyEnrollment(ctx context.Context, userID uint64, method, code string) error {
	if !u.enabled {
		return domain.ErrTfaDisabled
	}
	state, err := u.stateRepo.Get(ctx, domain.TfaPurposeEnrollment, userID, method)
	if err != nil {
		return err
	}
	if state == nil {
		return domain.ErrTfaChallengeNotFound
	}
	if err := u.validateCode(state, code); err != nil {
		return err
	}
	state.LastCode = code
	state.Verified = true
	if err := u.stateRepo.Put(ctx, state); err != nil {
		return err
	}
	return nil
}

func (u *tfaUsecase) CompleteEnrollment(ctx context.Context, userID uint64, method string) error {
	if !u.enabled {
		return domain.ErrTfaDisabled
	}
	state, err := u.stateRepo.Get(ctx, domain.TfaPurposeEnrollment, userID, method)
	if err != nil {
		return err
	}
	if state == nil {
		return domain.ErrTfaChallengeNotFound
	}
	if !state.Verified {
		return domain.ErrTfaNotVerified
	}
	if err := u.userRepo.SetTfaConfig(ctx, userID, state.Secret, method); err != nil {
		return err
	}
	_ = u.stateRepo.Delete(ctx, domain.TfaPurposeEnrollment, userID, method)
	return nil
}

func (u *tfaUsecase) RefreshChallenge(ctx context.Context, userID uint64, method string) (*dto.TfaRefreshResponse, error) {
	if !u.enabled {
		return nil, domain.ErrTfaDisabled
	}
	if method != domain.TfaMethodEmail {
		return nil, domain.ErrTfaMethodUnsupported
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, domain.ErrInvalidCredentials
	}

	purpose := domain.TfaPurposeEnrollment
	if user.TFAMethod == domain.TfaMethodEmail {
		purpose = domain.TfaPurposeLogin
	}

	state, err := u.stateRepo.Get(ctx, purpose, userID, domain.TfaMethodEmail)
	if err != nil {
		return nil, err
	}
	if state != nil && !state.CooldownAt.IsZero() && time.Now().Before(state.CooldownAt) {
		return &dto.TfaRefreshResponse{
			EmailSent:     false,
			ExpiresAt:     state.ExpiresAt,
			CooldownUntil: state.CooldownAt,
		}, domain.ErrTfaCooldown
	}

	code, err := newSixDigitCode()
	if err != nil {
		return nil, err
	}
	ttl := tfaChallengeTTL
	if purpose == domain.TfaPurposeLogin {
		ttl = tfaLoginChallengeTTL
	}
	newState := &domain.TfaSetupState{
		UserID:     userID,
		Purpose:    purpose,
		Method:     domain.TfaMethodEmail,
		Secret:     code,
		ExpiresAt:  time.Now().Add(ttl),
		CooldownAt: time.Now().Add(tfaEmailResendCool),
	}
	if err := u.stateRepo.Put(ctx, newState); err != nil {
		return nil, err
	}
	if err := u.mailer.SendTfaCode(ctx, user.Email, user.FirstName, code); err != nil {
		return nil, err
	}
	return &dto.TfaRefreshResponse{
		EmailSent:     true,
		ExpiresAt:     newState.ExpiresAt,
		CooldownUntil: newState.CooldownAt,
	}, nil
}

func (u *tfaUsecase) VerifyLoginCode(ctx context.Context, input dto.TfaVerifyCodeRequest, lc connectors.LoginContext) (*dto.LoginResponse, error) {
	if !u.enabled {
		return nil, domain.ErrTfaDisabled
	}
	claims, err := u.preAuth.Verify(input.PreAuthToken)
	if err != nil {
		return nil, domain.ErrTfaInvalidPreAuth
	}
	userID, err := claims.UserID()
	if err != nil {
		return nil, domain.ErrTfaInvalidPreAuth
	}
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil || !user.Activated {
		return nil, domain.ErrInvalidCredentials
	}
	// If the user has enrolled a specific method, the pre-auth claim must match it.
	if user.TFAMethod != "" && user.TFAMethod != claims.Method {
		return nil, domain.ErrTfaMethodMismatch
	}
	// Un-enrolled users only get the EMAIL fallback (matches what Login issues).
	if user.TFAMethod == "" && claims.Method != domain.TfaMethodEmail {
		return nil, domain.ErrTfaMethodMismatch
	}

	switch claims.Method {
	case domain.TfaMethodTotp:
		if user.TFASecret == "" {
			return nil, domain.ErrTfaInvalidCode
		}
		if !totp.Validate(input.Code, user.TFASecret) {
			return nil, domain.ErrTfaInvalidCode
		}
	case domain.TfaMethodEmail:
		state, err := u.stateRepo.Get(ctx, domain.TfaPurposeLogin, user.ID, domain.TfaMethodEmail)
		if err != nil {
			return nil, err
		}
		if state == nil {
			return nil, domain.ErrTfaChallengeNotFound
		}
		if subtle.ConstantTimeCompare([]byte(state.Secret), []byte(input.Code)) != 1 {
			return nil, domain.ErrTfaInvalidCode
		}
		_ = u.stateRepo.Delete(ctx, domain.TfaPurposeLogin, user.ID, domain.TfaMethodEmail)
	default:
		return nil, domain.ErrTfaMethodUnsupported
	}

	pair, err := IssueTokenPair(ctx, u.userRepo, u.refreshRepo, u.signer, u.refreshTTL, user, lc)
	if err != nil {
		return nil, err
	}
	return &dto.LoginResponse{TokenPair: *pair, User: dto.ToUserResponse(*user)}, nil
}

func (u *tfaUsecase) UnifiedEnrollment(ctx context.Context, userID uint64, input dto.TfaEnrollmentRequest) (*dto.TfaEnrollmentResponse, error) {
	if !u.enabled {
		return nil, domain.ErrTfaDisabled
	}
	switch input.Stage {
	case domain.TfaStageInit:
		init, err := u.InitEnrollment(ctx, userID, input.Method)
		if err != nil {
			return nil, err
		}
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Init: init}, nil
	case domain.TfaStageVerify:
		if input.Code == "" {
			return nil, domain.ErrTfaInvalidCode
		}
		if err := u.VerifyEnrollment(ctx, userID, input.Method, input.Code); err != nil {
			return nil, err
		}
		verified := true
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Verified: &verified}, nil
	case domain.TfaStageComplete:
		if err := u.CompleteEnrollment(ctx, userID, input.Method); err != nil {
			return nil, err
		}
		enabled := true
		return &dto.TfaEnrollmentResponse{Stage: input.Stage, Enabled: &enabled}, nil
	default:
		return nil, domain.ErrTfaMethodUnsupported
	}
}

func (u *tfaUsecase) IssueLoginEmailChallenge(ctx context.Context, userID uint64, email, firstName string) error {
	if !u.enabled {
		return domain.ErrTfaDisabled
	}
	if email == "" {
		return domain.ErrTfaMethodUnsupported
	}
	code, err := newSixDigitCode()
	if err != nil {
		return err
	}
	state := &domain.TfaSetupState{
		UserID:     userID,
		Purpose:    domain.TfaPurposeLogin,
		Method:     domain.TfaMethodEmail,
		Secret:     code,
		ExpiresAt:  time.Now().Add(tfaLoginChallengeTTL),
		CooldownAt: time.Now().Add(tfaEmailResendCool),
	}
	if err := u.stateRepo.Put(ctx, state); err != nil {
		return err
	}
	if err := u.mailer.SendTfaCode(ctx, email, firstName, code); err != nil {
		return err
	}
	return nil
}

func (u *tfaUsecase) validateCode(state *domain.TfaSetupState, code string) error {
	switch state.Method {
	case domain.TfaMethodTotp:
		if state.LastCode != "" && state.LastCode == code {
			return domain.ErrTfaInvalidCode
		}
		if !totp.Validate(code, state.Secret) {
			return domain.ErrTfaInvalidCode
		}
		return nil
	case domain.TfaMethodEmail:
		if subtle.ConstantTimeCompare([]byte(state.Secret), []byte(code)) != 1 {
			return domain.ErrTfaInvalidCode
		}
		return nil
	default:
		return domain.ErrTfaMethodUnsupported
	}
}

func tfaAccountName(u *domain.User) string {
	if u.Email != "" {
		return u.Email
	}
	return u.Login
}

func newSixDigitCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func renderQRDataURL(otpURL string) (string, error) {
	key, err := otp.NewKeyFromURL(otpURL)
	if err != nil {
		return "", err
	}
	img, err := key.Image(200, 200)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
