package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
)

type fakeFactorRepo struct {
	byUser map[uuid.UUID][]domain.TfaFactor
}

func (f *fakeFactorRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.TfaFactor, error) {
	for _, factors := range f.byUser {
		for i := range factors {
			if factors[i].ID == id {
				return &factors[i], nil
			}
		}
	}
	return nil, nil
}

func (f *fakeFactorRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]domain.TfaFactor, error) {
	return f.byUser[userID], nil
}

func (f *fakeFactorRepo) FindByUserAndType(context.Context, uuid.UUID, domain.TfaFactorType) (*domain.TfaFactor, error) {
	return nil, nil
}

func (f *fakeFactorRepo) Create(context.Context, *domain.TfaFactor) error { return nil }
func (f *fakeFactorRepo) Confirm(context.Context, uuid.UUID, time.Time) error { return nil }

func (f *fakeFactorRepo) MarkUsed(_ context.Context, id uuid.UUID, at time.Time) error {
	for userID, factors := range f.byUser {
		for i := range factors {
			if factors[i].ID == id {
				at := at
				f.byUser[userID][i].LastUsedAt = &at
				return nil
			}
		}
	}
	return nil
}

func (f *fakeFactorRepo) Delete(context.Context, uuid.UUID) error            { return nil }
func (f *fakeFactorRepo) DeleteByUser(context.Context, uuid.UUID) error      { return nil }
func (f *fakeFactorRepo) DeleteUnconfirmed(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func newTfaTestUsecase(t *testing.T, user *domain.User, factorRepo *fakeFactorRepo, limiter *ratelimit.LoginLimiter) (*tfaUsecase, *jwt.PreAuthSigner, *fakeChallengeRepo) {
	t.Helper()
	challengeRepo := &fakeChallengeRepo{byHash: map[string]*domain.UserChallenge{}}
	preAuth := jwt.NewPreAuthSigner("test-secret", time.Hour)
	uc := NewTfaUsecase(
		&fakeUserRepo{user: user}, noopRefreshRepo{}, factorRepo, challengeRepo,
		nil, jwt.NewSigner("test-secret", time.Hour), preAuth, time.Hour, nil, limiter,
	).(*tfaUsecase)
	return uc, preAuth, challengeRepo
}

// ACC-F6 (replay): the same TOTP code must not authenticate twice, even
// though it stays cryptographically valid for its whole ~90s window.
func TestVerifyLoginCodeRejectsReplayedTOTP(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	userID := uuid.New()
	user := &domain.User{ID: userID, Status: domain.UserStatusActive}
	confirmedAt := time.Now().UTC()
	factorID := uuid.New()
	factorRepo := &fakeFactorRepo{byUser: map[uuid.UUID][]domain.TfaFactor{
		userID: {{ID: factorID, UserID: userID, Type: domain.TfaFactorTotp, Secret: secret, ConfirmedAt: &confirmedAt}},
	}}
	uc, preAuth, _ := newTfaTestUsecase(t, user, factorRepo, nil)

	token, _, err := preAuth.Sign(userID, factorID, string(domain.TfaFactorTotp))
	if err != nil {
		t.Fatalf("sign preauth: %v", err)
	}
	code, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}

	req := dto.TfaVerifyCodeRequest{PreAuthToken: token, Code: code}
	if _, err := uc.VerifyLoginCode(context.Background(), req, connectors.LoginContext{}); err != nil {
		t.Fatalf("first use: unexpected err %v", err)
	}
	if _, err := uc.VerifyLoginCode(context.Background(), req, connectors.LoginContext{}); !errors.Is(err, domain.ErrTfaInvalidCode) {
		t.Fatalf("replayed code: got %v, want ErrTfaInvalidCode", err)
	}
}

// ACC-F6 (brute force): wrong codes must lock the account out instead of
// allowing an unlimited number of guesses against the same pre-auth token.
func TestVerifyLoginCodeIsRateLimited(t *testing.T) {
	secret := "JBSWY3DPEHPK3PXP"
	userID := uuid.New()
	user := &domain.User{ID: userID, Status: domain.UserStatusActive}
	confirmedAt := time.Now().UTC()
	factorID := uuid.New()
	factorRepo := &fakeFactorRepo{byUser: map[uuid.UUID][]domain.TfaFactor{
		userID: {{ID: factorID, UserID: userID, Type: domain.TfaFactorTotp, Secret: secret, ConfirmedAt: &confirmedAt}},
	}}
	limiter := ratelimit.NewLoginLimiter(3, time.Hour, time.Hour)
	uc, preAuth, _ := newTfaTestUsecase(t, user, factorRepo, limiter)

	token, _, err := preAuth.Sign(userID, factorID, string(domain.TfaFactorTotp))
	if err != nil {
		t.Fatalf("sign preauth: %v", err)
	}

	for i := 0; i < 3; i++ {
		req := dto.TfaVerifyCodeRequest{PreAuthToken: token, Code: "000000"}
		if _, err := uc.VerifyLoginCode(context.Background(), req, connectors.LoginContext{}); !errors.Is(err, domain.ErrTfaInvalidCode) {
			t.Fatalf("attempt %d: got %v, want ErrTfaInvalidCode", i, err)
		}
	}

	correct, err := totp.GenerateCode(secret, time.Now())
	if err != nil {
		t.Fatalf("generate code: %v", err)
	}
	req := dto.TfaVerifyCodeRequest{PreAuthToken: token, Code: correct}
	if _, err := uc.VerifyLoginCode(context.Background(), req, connectors.LoginContext{}); !errors.Is(err, domain.ErrLoginBlocked) {
		t.Fatalf("after lockout, even the correct code: got %v, want ErrLoginBlocked", err)
	}
}
