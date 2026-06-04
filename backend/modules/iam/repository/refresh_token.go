package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"gorm.io/gorm"
)

type pgRefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) connectors.RefreshTokenRepository {
	return &pgRefreshTokenRepository{db: db}
}

func (r *pgRefreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *pgRefreshTokenRepository) FindActiveByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hash, time.Now().UTC()).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *pgRefreshTokenRepository) Revoke(ctx context.Context, id uint64) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", &now).Error
}

func (r *pgRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uint64) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *pgRefreshTokenRepository) FindActiveByID(ctx context.Context, id uint64) (*domain.RefreshToken, error) {
	var t domain.RefreshToken
	err := r.db.WithContext(ctx).
		Where("id = ? AND revoked_at IS NULL AND expires_at > ?", id, time.Now().UTC()).
		First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *pgRefreshTokenRepository) ListActiveByUser(ctx context.Context, userID uint64) ([]domain.RefreshToken, error) {
	var tokens []domain.RefreshToken
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now().UTC()).
		Order("created_at DESC").
		Find(&tokens).Error
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *pgRefreshTokenRepository) RevokeAllForUserExcept(ctx context.Context, userID uint64, exceptID uint64) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, exceptID).
		Update("revoked_at", &now).Error
}
