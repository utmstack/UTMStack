package repository

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	mail_connectors "github.com/utmstack/utmstack/backend/internal/mail/connectors"
	appconfig_connectors "github.com/utmstack/utmstack/backend/modules/appconfig/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/constants/templates"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgChallengeRepository struct{ db *gorm.DB }

func NewChallengeRepository(db *gorm.DB) connectors.ChallengeRepository {
	return &pgChallengeRepository{db: db}
}

func (r *pgChallengeRepository) Get(ctx context.Context, purpose domain.ChallengePurpose, userID uuid.UUID) (*domain.UserChallenge, error) {
	return r.findOne(ctx, "purpose = ? AND user_id = ?", purpose, userID)
}

func (r *pgChallengeRepository) FindBySecret(ctx context.Context, purpose domain.ChallengePurpose, secretHash string) (*domain.UserChallenge, error) {
	if secretHash == "" {
		return nil, nil
	}
	return r.findOne(ctx, "purpose = ? AND secret = ?", purpose, secretHash)
}

func (r *pgChallengeRepository) findOne(ctx context.Context, query string, args ...any) (*domain.UserChallenge, error) {
	var c domain.UserChallenge
	err := r.db.WithContext(ctx).Where(query, args...).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgChallengeRepository) Put(ctx context.Context, c *domain.UserChallenge) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "purpose"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"factor_id", "secret", "last_code", "expires_at", "cooldown_at",
		}),
	}).Create(c).Error
}

func (r *pgChallengeRepository) Delete(ctx context.Context, purpose domain.ChallengePurpose, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("purpose = ? AND user_id = ?", purpose, userID).
		Delete(&domain.UserChallenge{}).Error
}

func (r *pgChallengeRepository) DeleteExpired(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&domain.UserChallenge{})
	return res.RowsAffected, res.Error
}

type challengeMailer struct {
	mail       mail_connectors.MailService
	configRepo mail_connectors.MailConfigurationRepository
	brand      appconfig_connectors.BrandNameProvider
}

func NewChallengeMailer(
	mail mail_connectors.MailService,
	configRepo mail_connectors.MailConfigurationRepository,
	brand appconfig_connectors.BrandNameProvider,
) connectors.ChallengeMailer {
	return &challengeMailer{mail: mail, configRepo: configRepo, brand: brand}
}

func (m *challengeMailer) Send(ctx context.Context, purpose domain.ChallengePurpose, to, name, secret string) error {
	switch purpose {
	case domain.ChallengeActivation:
		return m.sendLink(ctx, to, name, secret, templates.UserInvitation, "InviteURL",
			m.brand.ProductName(ctx)+": activate your account")
	case domain.ChallengePasswordReset:
		return m.sendLink(ctx, to, name, secret, templates.ResetPassword, "ResetURL",
			m.brand.ProductName(ctx)+": reset your password")
	case domain.ChallengeTfaLogin, domain.ChallengeTfaEnrollment:
		return m.mail.SendTemplateMail(ctx, []string{to}, m.brand.ProductName(ctx)+": your verification code", templates.TfaCode, map[string]string{
			"FirstName":   name,
			"Code":        secret,
			"ProductName": m.brand.ProductName(ctx),
		}, "")
	}
	return fmt.Errorf("no mail defined for challenge purpose %q", purpose)
}

func (m *challengeMailer) sendLink(ctx context.Context, to, name, token, template, urlKey, subject string) error {
	cfg, err := m.configRepo.GetMailConfiguration(ctx)
	if err != nil {
		return fmt.Errorf("load mail config: %w", err)
	}
	if cfg.BaseUrl == "" {
		return fmt.Errorf("mail base url is not configured")
	}
	return m.mail.SendTemplateMail(ctx, []string{to}, subject, template, map[string]string{
		"FirstName":   name,
		urlKey:        fmt.Sprintf("%s/reset/finish?key=%s", strings.TrimRight(cfg.BaseUrl, "/"), url.QueryEscape(token)),
		"ProductName": m.brand.ProductName(ctx),
	}, "")
}
