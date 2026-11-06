package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
)

// fakeTfaUsecase reports a fixed HasConfirmedFactor answer and records whether
// a login challenge was actually issued.
type fakeTfaUsecase struct {
	confirmed      bool
	challengeIssued bool
}

func (f *fakeTfaUsecase) Enroll(context.Context, uuid.UUID, dto.TfaEnrollmentRequest) (*dto.TfaEnrollmentResponse, error) {
	return nil, nil
}
func (f *fakeTfaUsecase) VerifyLoginCode(context.Context, dto.TfaVerifyCodeRequest, connectors.LoginContext) (*dto.LoginResponse, error) {
	return nil, nil
}
func (f *fakeTfaUsecase) Disable(context.Context, uuid.UUID, string) error  { return nil }
func (f *fakeTfaUsecase) ResetForUser(context.Context, uuid.UUID) error    { return nil }
func (f *fakeTfaUsecase) IssueLoginChallenge(context.Context, *domain.User) (domain.TfaFactorType, error) {
	f.challengeIssued = true
	return domain.TfaFactorTotp, nil
}
func (f *fakeTfaUsecase) HasConfirmedFactor(context.Context, uuid.UUID) (bool, error) {
	return f.confirmed, nil
}

func loginUser(t *testing.T, password string) *domain.User {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	hashStr := string(hash)
	return &domain.User{ID: uuid.New(), Email: "user@utmstack.com", PasswordHash: &hashStr, Status: domain.UserStatusActive}
}

// ACC-F5: a confirmed TFA factor must never be skipped by the dev/rc-only
// global APP_TFA_ENABLED=false kill switch — only an account with NOTHING
// confirmed may skip straight to a full session.
func TestLoginAlwaysRequiresTFAForAConfirmedFactorEvenWhenGloballyDisabled(t *testing.T) {
	user := loginUser(t, "correct-horse")
	tfa := &fakeTfaUsecase{confirmed: true}
	preAuth := jwt.NewPreAuthSigner("test-secret", 5*time.Minute)

	u := NewAuthUsecase(
		&fakeUserRepo{user: user}, nil, noopRefreshRepo{}, nil,
		jwt.NewSigner("test-secret", time.Hour), nil, time.Hour, nil,
		tfa, nil, preAuth,
		false, // APP_TFA_ENABLED=false, as installer/docker/compose.go sets for dev/rc
	)

	resp, err := u.Login(context.Background(), dto.LoginRequest{Login: user.Email, Password: "correct-horse"}, connectors.LoginContext{})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if !resp.TfaRequired {
		t.Fatal("a confirmed factor must still require TFA even with the global flag disabled")
	}
	if resp.TokenPair.AccessToken != "" {
		t.Fatal("a full session must not be issued before the second factor is verified")
	}
	if !tfa.challengeIssued {
		t.Fatal("IssueLoginChallenge must have been called")
	}
}

// No regression: an account with no confirmed factor must still log in
// directly regardless of the global flag's value.
func TestLoginSkipsTFAWhenNoFactorIsConfirmed(t *testing.T) {
	user := loginUser(t, "correct-horse")
	tfa := &fakeTfaUsecase{confirmed: false}
	preAuth := jwt.NewPreAuthSigner("test-secret", 5*time.Minute)

	u := NewAuthUsecase(
		&fakeUserRepo{user: user}, nil, noopRefreshRepo{}, nil,
		jwt.NewSigner("test-secret", time.Hour), nil, time.Hour, nil,
		tfa, nil, preAuth,
		false,
	)

	resp, err := u.Login(context.Background(), dto.LoginRequest{Login: user.Email, Password: "correct-horse"}, connectors.LoginContext{})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.TfaRequired {
		t.Fatal("an account with no confirmed factor must not be asked for one")
	}
	if resp.TokenPair.AccessToken == "" {
		t.Fatal("a full session must be issued when no factor is confirmed")
	}
	if tfa.challengeIssued {
		t.Fatal("IssueLoginChallenge must not have been called")
	}
}
