package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"gorm.io/gorm"
)

type pgAPIKeyRepository struct{ db *gorm.DB }

func NewAPIKeyRepository(db *gorm.DB) connectors.APIKeyRepository {
	return &pgAPIKeyRepository{db: db}
}

func (r *pgAPIKeyRepository) Create(ctx context.Context, k *domain.APIKey) error {
	return r.db.WithContext(ctx).Create(k).Error
}

func (r *pgAPIKeyRepository) Update(ctx context.Context, k *domain.APIKey) error {
	return r.db.WithContext(ctx).Save(k).Error
}

func (r *pgAPIKeyRepository) findOne(ctx context.Context, query string, args ...any) (*domain.APIKey, error) {
	var k domain.APIKey
	err := r.db.WithContext(ctx).Where(query, args...).First(&k).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}

func (r *pgAPIKeyRepository) FindByIDAndUser(ctx context.Context, id, userID uuid.UUID) (*domain.APIKey, error) {
	return r.findOne(ctx, "id = ? AND user_id = ?", id, userID)
}

func (r *pgAPIKeyRepository) FindByNameAndUser(ctx context.Context, name string, userID uuid.UUID) (*domain.APIKey, error) {
	return r.findOne(ctx, "name = ? AND user_id = ?", name, userID)
}

func (r *pgAPIKeyRepository) FindByHash(ctx context.Context, keyHash string) (*domain.APIKey, error) {
	if keyHash == "" {
		return nil, nil
	}
	return r.findOne(ctx, "key_hash = ?", keyHash)
}

func (r *pgAPIKeyRepository) DeleteByIDAndUser(ctx context.Context, id, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&domain.APIKey{}).Error
}

func (r *pgAPIKeyRepository) ListByUser(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.APIKey, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.APIKey{}).Where("user_id = ?", userID)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var keys []domain.APIKey
	if err := q.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}
	return keys, total, nil
}
