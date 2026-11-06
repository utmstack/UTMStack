package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"github.com/utmstack/utmstack/backend/pkg/authz"
	"github.com/utmstack/utmstack/backend/pkg/jwt"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

type fakeChallengeRepo struct {
	byHash      map[string]*domain.UserChallenge
	findSecretCtx context.Context
	deleted     bool
	deleteCtx   context.Context
}

func (f *fakeChallengeRepo) Get(context.Context, domain.ChallengePurpose, uuid.UUID) (*domain.UserChallenge, error) {
	return nil, nil
}
func (f *fakeChallengeRepo) FindBySecret(ctx context.Context, purpose domain.ChallengePurpose, hash string) (*domain.UserChallenge, error) {
	f.findSecretCtx = ctx
	if c, ok := f.byHash[hash]; ok && c.Purpose == purpose {
		return c, nil
	}
	return nil, nil
}
func (f *fakeChallengeRepo) Put(context.Context, *domain.UserChallenge) error { return nil }
func (f *fakeChallengeRepo) Delete(ctx context.Context, _ domain.ChallengePurpose, _ uuid.UUID) error {
	f.deleted = true
	f.deleteCtx = ctx
	return nil
}
func (f *fakeChallengeRepo) DeleteExpired(context.Context, time.Time) (int64, error) { return 0, nil }

// Extends fakeUserRepo (defined in auth_empty_login_test.go) to record the ctx
// each call received, so the test can check which tenant it carried.
type tenantCapturingUserRepo struct {
	fakeUserRepo
	findByIDCtx context.Context
}

func (f *tenantCapturingUserRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	f.findByIDCtx = ctx
	return f.fakeUserRepo.FindByID(ctx, id)
}

type noopRefreshRepo struct{}

func (noopRefreshRepo) Create(context.Context, *domain.RefreshToken) error { return nil }
func (noopRefreshRepo) FindActiveByHash(context.Context, string) (*domain.RefreshToken, error) {
	return nil, nil
}
func (noopRefreshRepo) FindActiveByID(context.Context, uuid.UUID) (*domain.RefreshToken, error) {
	return nil, nil
}
func (noopRefreshRepo) ListActiveByUser(context.Context, uuid.UUID) ([]domain.RefreshToken, error) {
	return nil, nil
}
func (noopRefreshRepo) Revoke(context.Context, uuid.UUID) error             { return nil }
func (noopRefreshRepo) RevokeAllForUser(context.Context, uuid.UUID) error   { return nil }
func (noopRefreshRepo) RevokeAllForUserExcept(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (noopRefreshRepo) RevokeOldestBeyond(context.Context, uuid.UUID, int) (int64, error) {
	return 0, nil
}
func (noopRefreshRepo) DeleteSpent(context.Context, time.Time, time.Time, int) (int64, error) {
	return 0, nil
}

// ACC-F30: a tenant's invite/reset link can carry the platform's own address
// (ACC-F33's fallback), so the request that finishes it resolves, via the Host
// header, to the platform tenant — not the tenant that issued the challenge.
// The key must still work: it is itself the authorization, and the rest of the
// flow must act in the challenge's real tenant, not the host-resolved one.
func TestSetPasswordFromChallengeWorksAcrossHostResolvedTenant(t *testing.T) {
	realTenant := uuid.New()
	platformTenant := uuid.New() // what the link's host incorrectly resolved to
	user := &domain.User{ID: uuid.New(), TenantID: realTenant, Status: domain.UserStatusActive}

	challenge := &domain.UserChallenge{
		UserID:    user.ID,
		TenantID:  realTenant,
		Purpose:   domain.ChallengePasswordReset,
		Secret:    secret.HashSHA256("the-token"),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}
	challengeRepo := &fakeChallengeRepo{byHash: map[string]*domain.UserChallenge{challenge.Secret: challenge}}
	userRepo := &tenantCapturingUserRepo{fakeUserRepo: fakeUserRepo{user: user}}

	u := NewAuthUsecase(
		userRepo, nil, noopRefreshRepo{}, challengeRepo,
		jwt.NewSigner("test-secret", time.Hour),
		nil, time.Hour, nil, nil, nil, nil, false,
	)

	// The request arrived on the platform's own hostname, so ResolveTenant
	// (backend/pkg/http/middleware/tenant.go) stamped the platform tenant.
	ctx := authz.WithTenantID(context.Background(), platformTenant.String())

	if err := u.SetPasswordFromChallenge(ctx, dto.ResetPasswordFinishRequest{Key: "the-token", NewPassword: "new-correct-password"}); err != nil {
		t.Fatalf("SetPasswordFromChallenge: %v (key must work regardless of which tenant the host resolved to)", err)
	}

	if !tenancy.ReadsAllTenants(challengeRepo.findSecretCtx) {
		t.Fatalf("FindBySecret must be called with an all-tenant-read context")
	}
	if got := authz.TenantIDFromContext(userRepo.findByIDCtx); got != realTenant.String() {
		t.Fatalf("FindByID tenant = %q, want the challenge's real tenant %q (not the host-resolved %q)", got, realTenant, platformTenant)
	}
	if !challengeRepo.deleted {
		t.Fatalf("the challenge must be deleted once consumed")
	}
	if got := authz.TenantIDFromContext(challengeRepo.deleteCtx); got != realTenant.String() {
		t.Fatalf("Delete tenant = %q, want the challenge's real tenant %q", got, realTenant)
	}
}
