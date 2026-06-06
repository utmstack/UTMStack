package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/modules/logstash/dto"
	"gorm.io/gorm"
)

type pgFilterGroupRepository struct {
	db *gorm.DB
}

func NewFilterGroupRepository(db *gorm.DB) connectors.FilterGroupRepository {
	return &pgFilterGroupRepository{db: db}
}

func (r *pgFilterGroupRepository) Create(ctx context.Context, group *domain.UtmLogstashFilterGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

func (r *pgFilterGroupRepository) Update(ctx context.Context, group *domain.UtmLogstashFilterGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

func (r *pgFilterGroupRepository) GetByID(ctx context.Context, id int64) (*domain.UtmLogstashFilterGroup, error) {
	var group domain.UtmLogstashFilterGroup
	err := r.db.WithContext(ctx).First(&group, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &group, nil
}

func (r *pgFilterGroupRepository) List(ctx context.Context, filters dto.FilterGroupListFilters) ([]domain.UtmLogstashFilterGroup, int64, error) {
	q := r.applyListWhere(r.db.WithContext(ctx).Model(&domain.UtmLogstashFilterGroup{}), filters)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var groups []domain.UtmLogstashFilterGroup
	offset := filters.Page * filters.Size
	if err := q.Offset(offset).Limit(filters.Size).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

func (r *pgFilterGroupRepository) applyListWhere(q *gorm.DB, f dto.FilterGroupListFilters) *gorm.DB {
	if f.IDEquals != nil {
		q = q.Where("id = ?", *f.IDEquals)
	}
	if f.GroupNameContains != nil {
		q = q.Where("group_name ILIKE ?", "%"+*f.GroupNameContains+"%")
	}
	if f.GroupDescriptionContains != nil {
		q = q.Where("group_description ILIKE ?", "%"+*f.GroupDescriptionContains+"%")
	}
	return q
}

func (r *pgFilterGroupRepository) Count(ctx context.Context, filters dto.FilterGroupCountFilters) (int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.UtmLogstashFilterGroup{})

	if filters.IDEquals != nil {
		q = q.Where("id = ?", *filters.IDEquals)
	}
	if filters.GroupNameContains != nil {
		q = q.Where("group_name ILIKE ?", "%"+*filters.GroupNameContains+"%")
	}
	if filters.GroupDescriptionContains != nil {
		q = q.Where("group_description ILIKE ?", "%"+*filters.GroupDescriptionContains+"%")
	}

	var count int64
	return count, q.Count(&count).Error
}

func (r *pgFilterGroupRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmLogstashFilterGroup{}, "id = ?", id).Error
}
