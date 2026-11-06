package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"github.com/utmstack/utmstack/backend/pkg/tenancy"
)

const (
	activationTTL = 7 * 24 * time.Hour
	resetTTL      = 1 * time.Hour
)

func issueLinkChallenge(
	ctx context.Context,
	repo connectors.ChallengeRepository,
	mailer connectors.ChallengeMailer,
	u *domain.User,
	purpose domain.ChallengePurpose,
	ttl time.Duration,
) error {
	token, err := secret.GenerateOpaque()
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	c := &domain.UserChallenge{
		UserID:    u.ID,
		TenantID:  u.TenantID,
		Purpose:   purpose,
		Secret:    secret.HashSHA256(token),
		ExpiresAt: now.Add(ttl),
	}
	if err := repo.Put(ctx, c); err != nil {
		return err
	}
	if mailer == nil {
		return nil
	}
	return mailer.Send(ctx, purpose, u.Email, u.Name, token)
}

// resolveLinkChallenge looks the token up across every tenant: the mail
// BaseUrl a tenant's invite/reset link uses can fall back to the platform's
// own address (ACC-F33), which the host-to-tenant resolver then maps to the
// platform tenant, not the one that issued the challenge (ACC-F30) — so
// scoping this read to whatever tenant the link's host happened to resolve to
// would refuse a perfectly valid key. The token is a random, SHA-256-hashed
// opaque secret: it is already the authorization, so an all-tenant read here
// costs nothing in security and the challenge's own TenantID is what the
// caller should use from here on.
func resolveLinkChallenge(ctx context.Context, repo connectors.ChallengeRepository, token string) (*domain.UserChallenge, error) {
	ctx = tenancy.WithAllTenantsRead(ctx)
	hash := secret.HashSHA256(token)
	for _, purpose := range []domain.ChallengePurpose{domain.ChallengePasswordReset, domain.ChallengeActivation} {
		c, err := repo.FindBySecret(ctx, purpose, hash)
		if err != nil {
			return nil, err
		}
		if c == nil {
			continue
		}
		if !c.ExpiresAt.After(time.Now().UTC()) {
			return nil, domain.ErrChallengeExpired
		}
		return c, nil
	}
	return nil, domain.ErrChallengeNotFound
}
