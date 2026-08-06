package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

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

func (r *pgRefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", &now).Error
}

func (r *pgRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", &now).Error
}

func (r *pgRefreshTokenRepository) FindActiveByID(ctx context.Context, id uuid.UUID) (*domain.RefreshToken, error) {
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

func (r *pgRefreshTokenRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]domain.RefreshToken, error) {
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

func (r *pgRefreshTokenRepository) RevokeAllForUserExcept(ctx context.Context, userID uuid.UUID, exceptID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, exceptID).
		Update("revoked_at", &now).Error
}

func (r *pgRefreshTokenRepository) RevokeOldestBeyond(ctx context.Context, userID uuid.UUID, keep int) (int64, error) {
	if keep <= 0 {
		return 0, nil
	}
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where(`user_id = ? AND revoked_at IS NULL AND expires_at > ? AND id NOT IN (
			SELECT id FROM refresh_tokens
			WHERE user_id = ? AND revoked_at IS NULL AND expires_at > ?
			ORDER BY created_at DESC LIMIT ?
		)`, userID, now, userID, now, keep).
		Update("revoked_at", &now)
	return res.RowsAffected, res.Error
}

// DeleteSpent keeps revoked rows for a grace period rather than removing them at
// once: "this session was closed at 03:14" is exactly what gets read after an
// incident, and it would be gone.
func (r *pgRefreshTokenRepository) DeleteSpent(ctx context.Context, graceCutoff, expiredCutoff time.Time, limit int) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("(revoked_at IS NOT NULL AND revoked_at < ?) OR expires_at < ?", graceCutoff, expiredCutoff).
		Limit(limit).
		Delete(&domain.RefreshToken{})
	return res.RowsAffected, res.Error
}
