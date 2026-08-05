package usecase

import (
	"context"
	"testing"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	keyTenant   = "8f1c1b8e-0000-4000-8000-000000000001"
	otherTenant = "8f1c1b8e-0000-4000-8000-000000000002"
)

type fakeKeyRepo struct {
	connectors.APIKeyRepository
	key          *domain.APIKey
	sawAllTenant bool
}

func (f *fakeKeyRepo) FindByKey(ctx context.Context, _ string) (*domain.APIKey, error) {
	f.sawAllTenant = tenancy.SpansAllTenants(ctx)
	return f.key, nil
}

type fakeKeyUserRepo struct {
	connectors.UserRepository
	user      *domain.User
	sawTenant string
}

func (f *fakeKeyUserRepo) FindByID(ctx context.Context, _ uint64) (*domain.User, error) {
	f.sawTenant = authz.TenantIDFromContext(ctx)
	return f.user, nil
}

func (f *fakeKeyUserRepo) FindPermissionsByUserID(context.Context, uint64) ([]string, error) {
	return []string{"alerts.read"}, nil
}

func (f *fakeKeyUserRepo) FindRolesByUserID(context.Context, uint64) ([]domain.Authority, error) {
	return []domain.Authority{{Name: "ROLE_USER"}}, nil
}

// A service that authenticates a pusher holds the internal key and no tenant.
// Scoping the lookup by the caller's tenant fails the query outright, so every
// key-authenticated push would be rejected on an MSSP install.
func TestAPIKeyAuthenticateHasNoTenantToScopeBy(t *testing.T) {
	keys := &fakeKeyRepo{key: &domain.APIKey{UserID: 7, TenantID: keyTenant}}
	users := &fakeKeyUserRepo{user: &domain.User{ID: 7, Login: "pusher", Activated: true, TenantID: keyTenant}}
	uc := NewAPIKeyUsecase(keys, users)

	res, err := uc.Authenticate(context.Background(), "some-key", "10.0.0.1")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if !keys.sawAllTenant {
		t.Error("the key lookup was scoped to a tenant the caller does not have")
	}
	if res.TenantID != keyTenant {
		t.Errorf("tenant = %q, want the key's tenant %q", res.TenantID, keyTenant)
	}
}

// The key says which tenant this is. A request that arrived on another tenant's
// host must not turn its key into that host's tenant; it authenticates as its
// own, and is turned away where tenants are enforced.
func TestAPIKeyAuthenticateIgnoresTheCallersTenant(t *testing.T) {
	keys := &fakeKeyRepo{key: &domain.APIKey{UserID: 7, TenantID: keyTenant}}
	users := &fakeKeyUserRepo{user: &domain.User{ID: 7, Login: "pusher", Activated: true, TenantID: keyTenant}}
	uc := NewAPIKeyUsecase(keys, users)

	ctx := authz.WithTenantID(context.Background(), otherTenant)
	res, err := uc.Authenticate(ctx, "some-key", "10.0.0.1")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if users.sawTenant != keyTenant {
		t.Errorf("the user was looked up in %q, want the key's tenant %q", users.sawTenant, keyTenant)
	}
	if res.TenantID != keyTenant {
		t.Errorf("tenant = %q, want %q", res.TenantID, keyTenant)
	}
}
