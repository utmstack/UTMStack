package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"gorm.io/gorm"
)

type pgAlertTagRepository struct {
	db *gorm.DB
}

func NewAlertTagRepository(db *gorm.DB) connectors.AlertTagRepository {
	return &pgAlertTagRepository{db: db}
}

func (r *pgAlertTagRepository) Create(ctx context.Context, tag domain.UtmAlertTag) (*domain.UtmAlertTag, error) {
	if err := r.db.WithContext(ctx).Create(&tag).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTagNameTaken
		}
		return nil, err
	}
	return &tag, nil
}

func (r *pgAlertTagRepository) Update(ctx context.Context, tag domain.UtmAlertTag) (*domain.UtmAlertTag, error) {
	if err := r.db.WithContext(ctx).Save(&tag).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrTagNameTaken
		}
		return nil, err
	}
	return &tag, nil
}

func (r *pgAlertTagRepository) List(ctx context.Context, f dto.AlertTagFilters) ([]domain.UtmAlertTag, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Size < 1 || f.Size > 200 {
		f.Size = 20
	}

	q := r.db.WithContext(ctx).Model(&domain.UtmAlertTag{})

	if f.TagName != nil && *f.TagName != "" {
		q = q.Where("tag_name ILIKE ?", "%"+*f.TagName+"%")
	}
	if f.SystemOwner != nil {
		q = q.Where("system_owner = ?", *f.SystemOwner)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.UtmAlertTag
	if err := q.
		Order("id ASC").
		Offset((f.Page - 1) * f.Size).
		Limit(f.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgAlertTagRepository) GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTag, error) {
	var row domain.UtmAlertTag
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgAlertTagRepository) FindByIDs(ctx context.Context, ids []int64) ([]domain.UtmAlertTag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []domain.UtmAlertTag
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgAlertTagRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmAlertTag{}, id).Error
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
