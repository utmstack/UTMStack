package usecase

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	BootstrapAdminLogin    = "admin"
	BootstrapAdminEmail    = "admin@localhost"
	bootstrapDefaultTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"
)

func (u *userUsecase) EnsureBootstrapAdmin(ctx context.Context, password string) (created bool, generated string, err error) {
	ctx = tenancy.WithAllTenants(ctx)

	n, err := u.userRepo.Count(ctx)
	if err != nil {
		return false, "", fmt.Errorf("counting users: %w", err)
	}
	if n > 0 {
		return false, "", nil
	}

	if password == "" {
		password, err = secret.GenerateOpaque()
		if err != nil {
			return false, "", fmt.Errorf("generating a password: %w", err)
		}
		generated = password
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, "", fmt.Errorf("hashing the password: %w", err)
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
		return false, "", fmt.Errorf("creating the bootstrap admin: %w", err)
	}
	roles := []string{AdminRoleName, authz.RolePlatformAdmin}
	if err := u.rbacRepo.ReplaceUserRoles(ctx, admin.ID, roles, "system"); err != nil {
		return false, "", fmt.Errorf("granting %v: %w", roles, err)
	}

	return true, generated, nil
}
