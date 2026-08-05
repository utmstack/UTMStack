package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	BootstrapAdminLogin    = "admin"
	BootstrapAdminEmail    = "admin@localhost"
	bootstrapDefaultTenant = authz.DefaultTenantID
)

// ErrBootstrapPasswordRequired means the install has no users and no password
// was supplied to create the first one with.
var ErrBootstrapPasswordRequired = errors.New(
	"UTMSTACK_ADMIN_PASSWORD is required to create the initial administrator")

func (u *userUsecase) EnsureBootstrapAdmin(ctx context.Context, password string) (created bool, err error) {
	ctx = tenancy.WithAllTenants(ctx)

	n, err := u.userRepo.Count(ctx)
	if err != nil {
		return false, fmt.Errorf("counting users: %w", err)
	}
	if n > 0 {
		return false, nil
	}

	// Checked only once we know an account has to be created, so an upgrade of
	// an install that already has users never needs it set.
	if password == "" {
		return false, ErrBootstrapPasswordRequired
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, fmt.Errorf("hashing the password: %w", err)
	}

	now := time.Now().UTC()
	admin := &domain.User{
		Login:        BootstrapAdminLogin,
		Email:        BootstrapAdminEmail,
		PasswordHash: string(hash),
		FirstName:    "Admin",
		LangKey:      "en",
		Activated:    true,

		// Marks the account in the user list as still holding the password it
		// was created with.
		DefaultPassword: true,

		// The installer looks the account up by created_by = 'system'.
		CreatedBy:   "system",
		CreatedDate: now,
		TenantID:    bootstrapDefaultTenant,
	}

	if err := u.userRepo.Create(ctx, admin); err != nil {
		return false, fmt.Errorf("creating the bootstrap admin: %w", err)
	}

	roles := []string{AdminRoleName}
	if err := u.rbacRepo.ReplaceUserRoles(ctx, admin.ID, roles, "system"); err != nil {
		return false, fmt.Errorf("granting %v: %w", roles, err)
	}

	return true, nil
}
