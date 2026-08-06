package usecase

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
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

func resolveLinkChallenge(ctx context.Context, repo connectors.ChallengeRepository, token string) (*domain.UserChallenge, error) {
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
