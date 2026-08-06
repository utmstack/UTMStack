package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/utmstack/utmstack/backend/modules/iam/connectors"
	"github.com/utmstack/utmstack/backend/modules/iam/domain"
	"github.com/utmstack/utmstack/backend/modules/iam/dto"
	"gorm.io/gorm"
)

type pgIdentityProviderRepository struct{ db *gorm.DB }

func NewIdentityProviderRepository(db *gorm.DB) connectors.IdentityProviderRepository {
	return &pgIdentityProviderRepository{db: db}
}

func (r *pgIdentityProviderRepository) Save(ctx context.Context, c *domain.IdentityProviderConfig) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *pgIdentityProviderRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.IdentityProviderConfig, error) {
	var c domain.IdentityProviderConfig
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgIdentityProviderRepository) FindByName(ctx context.Context, name string) (*domain.IdentityProviderConfig, error) {
	var c domain.IdentityProviderConfig
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *pgIdentityProviderRepository) List(ctx context.Context, f dto.IdentityProviderFilter) ([]domain.IdentityProviderConfig, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.IdentityProviderConfig{})
	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if f.ProviderType != "" {
		q = q.Where("provider_type = ?", f.ProviderType)
	}
	if f.Active != nil {
		q = q.Where("active = ?", *f.Active)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.IdentityProviderConfig
	if err := q.Order("name ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgIdentityProviderRepository) ListActive(ctx context.Context) ([]domain.IdentityProviderConfig, error) {
	var items []domain.IdentityProviderConfig
	if err := r.db.WithContext(ctx).Where("active = ?", true).Order("name ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgIdentityProviderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.IdentityProviderConfig{}).Error
}

func (r *pgIdentityProviderRepository) ListMappings(ctx context.Context, providerID uuid.UUID) ([]domain.IdentityProviderGroupMapping, error) {
	var rows []domain.IdentityProviderGroupMapping
	err := r.db.WithContext(ctx).
		Where("identity_provider_id = ?", providerID).
		Order("group_name").
		Find(&rows).Error
	return rows, err
}

// ReplaceMappings rewrites the whole set in one transaction. A mapping edited
// row by row would leave a window where a login sees half of it.
func (r *pgIdentityProviderRepository) ReplaceMappings(
	ctx context.Context, providerID uuid.UUID, mappings []domain.IdentityProviderGroupMapping,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("identity_provider_id = ?", providerID).
			Delete(&domain.IdentityProviderGroupMapping{}).Error; err != nil {
			return err
		}
		if len(mappings) == 0 {
			return nil
		}
		return tx.Create(&mappings).Error
	})
}
