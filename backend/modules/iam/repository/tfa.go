package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/pkg/secret"
	"gorm.io/gorm"
)

type pgTfaFactorRepository struct {
	db     *gorm.DB
	cipher *secret.Cipher
}

func NewTfaFactorRepository(db *gorm.DB, cipher *secret.Cipher) connectors.TfaFactorRepository {
	return &pgTfaFactorRepository{db: db, cipher: cipher}
}

func (r *pgTfaFactorRepository) seal(f *domain.TfaFactor) error {
	if f.Type != domain.TfaFactorTotp || f.Secret == "" {
		return nil
	}
	stored, err := encryptTfaSecret(r.cipher, f.Secret)
	if err != nil {
		return err
	}
	f.Secret = stored
	return nil
}

func (r *pgTfaFactorRepository) open(f *domain.TfaFactor) error {
	if f.Type != domain.TfaFactorTotp || f.Secret == "" {
		return nil
	}
	plain, err := decryptTfaSecret(r.cipher, f.Secret)
	if err != nil {
		return err
	}
	f.Secret = plain
	return nil
}

func (r *pgTfaFactorRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.TfaFactor, error) {
	return r.findOne(ctx, "id = ?", id)
}

func (r *pgTfaFactorRepository) FindByUserAndType(ctx context.Context, userID uuid.UUID, t domain.TfaFactorType) (*domain.TfaFactor, error) {
	return r.findOne(ctx, "user_id = ? AND type = ?", userID, t)
}

func (r *pgTfaFactorRepository) findOne(ctx context.Context, query string, args ...any) (*domain.TfaFactor, error) {
	var f domain.TfaFactor
	err := r.db.WithContext(ctx).Where(query, args...).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err := r.open(&f); err != nil {
		return nil, err
	}
	return &f, nil
}

func (r *pgTfaFactorRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.TfaFactor, error) {
	var factors []domain.TfaFactor
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("type, created_at").
		Find(&factors).Error
	if err != nil {
		return nil, err
	}
	for i := range factors {
		if err := r.open(&factors[i]); err != nil {
			return nil, err
		}
	}
	return factors, nil
}

func (r *pgTfaFactorRepository) Create(ctx context.Context, f *domain.TfaFactor) error {
	plain := f.Secret
	if err := r.seal(f); err != nil {
		return err
	}
	err := r.db.WithContext(ctx).Create(f).Error
	f.Secret = plain
	return err
}

func (r *pgTfaFactorRepository) Confirm(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.TfaFactor{}).
		Where("id = ?", id).
		Update("confirmed_at", at).Error
}

func (r *pgTfaFactorRepository) MarkUsed(ctx context.Context, id uuid.UUID, at time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.TfaFactor{}).
		Where("id = ?", id).
		Update("last_used_at", at).Error
}

func (r *pgTfaFactorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.TfaFactor{}).Error
}

func (r *pgTfaFactorRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&domain.TfaFactor{}).Error
}

func (r *pgTfaFactorRepository) DeleteUnconfirmed(ctx context.Context, before time.Time) (int64, error) {
	res := r.db.WithContext(ctx).
		Where("confirmed_at IS NULL AND created_at < ?", before).
		Delete(&domain.TfaFactor{})
	return res.RowsAffected, res.Error
}

const tfaEncPrefix = "enc:v1:"

func encryptTfaSecret(c *secret.Cipher, plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	crypt, err := c.Encrypt(plain)
	if err != nil {
		return "", err
	}
	return tfaEncPrefix + crypt, nil
}

func decryptTfaSecret(c *secret.Cipher, stored string) (string, error) {
	rest, tagged := strings.CutPrefix(stored, tfaEncPrefix)
	if !tagged {
		return stored, nil
	}
	return c.Decrypt(rest)
}
