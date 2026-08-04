package usecase

import (
	"context"
	"slices"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

// The embedded interfaces are left nil on purpose: a method the bootstrap is not
// supposed to call panics instead of quietly returning a zero value.
type fakeUserRepo struct {
	connectors.UserRepository
	count   int64
	created []*domain.User
}

func (f *fakeUserRepo) Count(context.Context) (int64, error) { return f.count, nil }

func (f *fakeUserRepo) Create(_ context.Context, u *domain.User) error {
	u.ID = uint64(len(f.created) + 1)
	f.created = append(f.created, u)
	f.count++
	return nil
}

type fakeRBACRepo struct {
	connectors.RBACRepository
	granted map[uint64][]string
}

func (f *fakeRBACRepo) ReplaceUserRoles(_ context.Context, userID uint64, roles []string, _ string) error {
	if f.granted == nil {
		f.granted = map[uint64][]string{}
	}
	f.granted[userID] = roles
	return nil
}

func newBootstrapUsecase(users *fakeUserRepo, rbac *fakeRBACRepo) *userUsecase {
	return &userUsecase{userRepo: users, rbacRepo: rbac}
}

func TestBootstrapCreatesTheFirstAdmin(t *testing.T) {
	users := &fakeUserRepo{}
	rbac := &fakeRBACRepo{}

	created, generated, err := newBootstrapUsecase(users, rbac).EnsureBootstrapAdmin(context.Background(), "")
	if err != nil {
		t.Fatalf("EnsureBootstrapAdmin: %v", err)
	}
	if !created {
		t.Fatal("no admin was created on an empty install")
	}
	if generated == "" {
		t.Fatal("no password was returned; the operator would have no way in")
	}
	if len(users.created) != 1 {
		t.Fatalf("created %d users, want 1", len(users.created))
	}

	admin := users.created[0]
	if admin.Login != BootstrapAdminLogin {
		t.Errorf("Login = %q, want %q", admin.Login, BootstrapAdminLogin)
	}
	// The frontend's first-run gate keys off this exact address.
	if admin.Email != BootstrapAdminEmail {
		t.Errorf("Email = %q, want %q", admin.Email, BootstrapAdminEmail)
	}
	if !admin.Activated {
		t.Error("the account is not activated, so nobody can sign in with it")
	}
	if !admin.DefaultPassword {
		t.Error("DefaultPassword is false; the user list would not flag the account")
	}
	// The installer looks the account up by this.
	if admin.CreatedBy != "system" {
		t.Errorf("CreatedBy = %q, want %q", admin.CreatedBy, "system")
	}
	if admin.TenantID != bootstrapDefaultTenant {
		t.Errorf("TenantID = %q, want the tenant every backfilled row uses", admin.TenantID)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(generated)); err != nil {
		t.Fatalf("the returned password does not match the stored hash: %v", err)
	}

	// Both roles: an instance is installed as Community and licensed later, so
	// without the platform role the operator loses tenant management the moment
	// they apply an MSSP licence.
	roles := rbac.granted[admin.ID]
	if !slices.Contains(roles, AdminRoleName) {
		t.Errorf("granted roles = %v, want it to include %s", roles, AdminRoleName)
	}
	if !slices.Contains(roles, authz.RolePlatformAdmin) {
		t.Errorf("granted roles = %v, want it to include %s", roles, authz.RolePlatformAdmin)
	}
}

// This runs on every boot, so a second start must not add a second admin.
func TestBootstrapIsIdempotent(t *testing.T) {
	users := &fakeUserRepo{}
	rbac := &fakeRBACRepo{}
	uc := newBootstrapUsecase(users, rbac)

	if _, _, err := uc.EnsureBootstrapAdmin(context.Background(), ""); err != nil {
		t.Fatalf("first boot: %v", err)
	}

	created, generated, err := uc.EnsureBootstrapAdmin(context.Background(), "")
	if err != nil {
		t.Fatalf("second boot: %v", err)
	}
	if created {
		t.Fatal("a second admin was created on the next boot")
	}
	if generated != "" {
		t.Fatalf("a password was returned without creating anything: %q", generated)
	}
	if len(users.created) != 1 {
		t.Fatalf("%d users exist, want 1", len(users.created))
	}
}

// An install that already has users — every upgrade from v11 — must be left
// completely alone.
func TestBootstrapSkipsAnInstallThatHasUsers(t *testing.T) {
	users := &fakeUserRepo{count: 3}
	rbac := &fakeRBACRepo{}

	created, _, err := newBootstrapUsecase(users, rbac).EnsureBootstrapAdmin(context.Background(), "")
	if err != nil {
		t.Fatalf("EnsureBootstrapAdmin: %v", err)
	}
	if created {
		t.Fatal("an admin was created on an install that already had users")
	}
	if len(users.created) != 0 {
		t.Fatalf("wrote %d users to an existing install", len(users.created))
	}
}

func TestBootstrapUsesTheSuppliedPassword(t *testing.T) {
	users := &fakeUserRepo{}
	rbac := &fakeRBACRepo{}

	const chosen = "a-password-the-installer-picked"
	created, generated, err := newBootstrapUsecase(users, rbac).EnsureBootstrapAdmin(context.Background(), chosen)
	if err != nil {
		t.Fatalf("EnsureBootstrapAdmin: %v", err)
	}
	if !created {
		t.Fatal("no admin was created")
	}
	// Nothing generated means nothing to log: the caller already knows it.
	if generated != "" {
		t.Fatalf("generated = %q, want empty when a password was supplied", generated)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(users.created[0].PasswordHash), []byte(chosen)); err != nil {
		t.Fatalf("the supplied password does not match the stored hash: %v", err)
	}
}
