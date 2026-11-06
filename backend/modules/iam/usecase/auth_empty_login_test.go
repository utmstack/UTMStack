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
	"github.com/utmstack/utmstack/backend/pkg/ratelimit"
)

type fakeUserRepo struct{ user *domain.User }

func (f *fakeUserRepo) FindByID(context.Context, uuid.UUID) (*domain.User, error) { return f.user, nil }
func (f *fakeUserRepo) FindByEmail(context.Context, string) (*domain.User, error) {
	return f.user, nil
}
func (f *fakeUserRepo) List(context.Context, connectors.ListUsersFilter) ([]domain.User, int64, error) {
	return nil, 0, nil
}
func (f *fakeUserRepo) Create(context.Context, *domain.User) error { return nil }
func (f *fakeUserRepo) Update(context.Context, *domain.User) error { return nil }
func (f *fakeUserRepo) UpdatePassword(context.Context, uuid.UUID, string) error { return nil }
func (f *fakeUserRepo) UpdateStatus(context.Context, uuid.UUID, domain.UserStatus) error {
	return nil
}
func (f *fakeUserRepo) Count(context.Context) (int64, error) { return 0, nil }
func (f *fakeUserRepo) ExistsByEmail(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (f *fakeUserRepo) FindPermissionsByUserID(context.Context, uuid.UUID) ([]string, error) {
	return nil, nil
}
func (f *fakeUserRepo) FindRolesByUserID(context.Context, uuid.UUID) ([]domain.Role, error) {
	return nil, nil
}
func (f *fakeUserRepo) FindRolesByUserIDs(context.Context, []uuid.UUID) (map[uuid.UUID][]domain.Role, error) {
	return nil, nil
}

// ACC-F12: EmptyLogin performs the same bcrypt compare as Login and must be
// covered by the same IP lockout — otherwise it is an unlimited-attempt
// password oracle reachable without ever going through Login.
func TestEmptyLoginIsRateLimited(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-horse"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	hashStr := string(hash)
	user := &domain.User{ID: uuid.New(), Email: "victim@utmstack.com", PasswordHash: &hashStr, Status: domain.UserStatusActive}

	limiter := ratelimit.NewLoginLimiter(3, time.Minute, time.Minute)
	u := NewAuthUsecase(
		&fakeUserRepo{user: user},
		nil, nil, nil,
		jwt.NewSigner("test-secret", time.Hour),
		limiter,
		time.Hour,
		nil, nil, nil, nil,
		false,
	)

	lc := connectors.LoginContext{IP: "198.51.100.5"}
	badInput := dto.LoginRequest{Login: "victim@utmstack.com", Password: "wrong-guess"}

	for i := 0; i < 3; i++ {
		if _, err := u.EmptyLogin(context.Background(), badInput, lc); err != domain.ErrInvalidCredentials {
			t.Fatalf("attempt %d: got %v, want ErrInvalidCredentials", i, err)
		}
	}

	// The 4th attempt — even with the correct password — must be refused by the
	// lockout, never reaching the bcrypt compare that would otherwise confirm it.
	goodInput := dto.LoginRequest{Login: "victim@utmstack.com", Password: "correct-horse"}
	if _, err := u.EmptyLogin(context.Background(), goodInput, lc); err != domain.ErrLoginBlocked {
		t.Fatalf("after 3 failures, got %v, want ErrLoginBlocked (unlimited attempts = password oracle)", err)
	}
}
