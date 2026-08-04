package repository

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Secret carries the TOTP secret during enrollment and the emailed code during
// login, so it is encrypted at rest here rather than in the usecase: this is the
// only pair of methods that reads and writes it, and encrypting further out
// would leave a place to forget.
type pgTfaStateRepository struct {
	db     *gorm.DB
	ttl    time.Duration
	cipher *secret.Cipher
}

func NewTfaStateRepository(db *gorm.DB, ttl time.Duration, cipher *secret.Cipher) connectors.TfaStateRepository {
	return &pgTfaStateRepository{db: db, ttl: ttl, cipher: cipher}
}

func (r *pgTfaStateRepository) Get(ctx context.Context, purpose string, userID uint64, method string) (*domain.TfaSetupState, error) {
	var row domain.TfaSetupState
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ? AND method = ?", userID, purpose, method).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !row.ExpiresAt.IsZero() && time.Now().After(row.ExpiresAt) {
		_ = r.Delete(ctx, purpose, userID, method)
		return nil, nil
	}
	plain, err := decryptTfaSecret(r.cipher, row.Secret)
	if err != nil {
		return nil, err
	}
	row.Secret = plain
	return &row, nil
}

func (r *pgTfaStateRepository) Put(ctx context.Context, state *domain.TfaSetupState) error {
	if state == nil {
		return nil
	}
	if state.ExpiresAt.IsZero() {
		state.ExpiresAt = time.Now().Add(r.ttl)
	}
	crypt, err := encryptTfaSecret(r.cipher, state.Secret)
	if err != nil {
		return err
	}

	// A copy, so the caller keeps the plaintext it passed in.
	stored := *state
	stored.Secret = crypt

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "purpose"}, {Name: "method"}},
		UpdateAll: true,
	}).Create(&stored).Error
}

func (r *pgTfaStateRepository) Delete(ctx context.Context, purpose string, userID uint64, method string) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND purpose = ? AND method = ?", userID, purpose, method).
		Delete(&domain.TfaSetupState{}).Error
}

func (r *pgTfaStateRepository) MarkVerified(ctx context.Context, purpose string, userID uint64, method string) error {
	return r.db.WithContext(ctx).Model(&domain.TfaSetupState{}).
		Where("user_id = ? AND purpose = ? AND method = ? AND (expires_at = ? OR expires_at > ?)",
			userID, purpose, method, time.Time{}, time.Now().UTC()).
		Update("verified", true).Error
}
