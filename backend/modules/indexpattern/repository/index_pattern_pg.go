package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/indexpattern/connectors"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/domain"
	"github.com/utmstack/utmstack/backend/modules/indexpattern/dto"
	"gorm.io/gorm"
)

type pgIndexPatternRepository struct {
	db *gorm.DB
}

func NewIndexPatternRepository(db *gorm.DB) connectors.IndexPatternRepository {
	return &pgIndexPatternRepository{db: db}
}

func (r *pgIndexPatternRepository) Create(ctx context.Context, p *domain.UtmIndexPattern) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *pgIndexPatternRepository) Update(ctx context.Context, p *domain.UtmIndexPattern) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *pgIndexPatternRepository) GetByID(ctx context.Context, id int64) (*domain.UtmIndexPattern, error) {
	var p domain.UtmIndexPattern
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgIndexPatternRepository) applyFilters(q *gorm.DB, f dto.IndexPatternFilters) *gorm.DB {
	if f.IDEquals != nil {
		q = q.Where("id = ?", *f.IDEquals)
	}
	if f.IDGreaterThan != nil {
		q = q.Where("id > ?", *f.IDGreaterThan)
	}
	if f.IDLessThan != nil {
		q = q.Where("id < ?", *f.IDLessThan)
	}
	if f.PatternEquals != nil {
		q = q.Where("pattern = ?", *f.PatternEquals)
	}
	if f.PatternContains != nil {
		q = q.Where("pattern ILIKE ?", "%"+*f.PatternContains+"%")
	}
	if f.ModuleEquals != nil {
		q = q.Where("pattern_module = ?", *f.ModuleEquals)
	}
	if f.ModuleContains != nil {
		q = q.Where("pattern_module ILIKE ?", "%"+*f.ModuleContains+"%")
	}
	if f.PatternSystem != nil {
		q = q.Where("pattern_system = ?", *f.PatternSystem)
	}
	if f.IsActive != nil {
		q = q.Where("is_active = ?", *f.IsActive)
	}
	return q
}

func (r *pgIndexPatternRepository) List(ctx context.Context, f dto.IndexPatternFilters) ([]domain.UtmIndexPattern, int64, error) {
	q := r.applyFilters(r.db.WithContext(ctx).Model(&domain.UtmIndexPattern{}), f)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	size := f.Size
	if size <= 0 {
		size = 20
	}
	offset := f.Page * size

	var items []domain.UtmIndexPattern
	if err := q.Order("id ASC").Offset(offset).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgIndexPatternRepository) Count(ctx context.Context, f dto.IndexPatternFilters) (int64, error) {
	q := r.applyFilters(r.db.WithContext(ctx).Model(&domain.UtmIndexPattern{}), f)
	var count int64
	return count, q.Count(&count).Error
}

func (r *pgIndexPatternRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmIndexPattern{}, id).Error
}

func (r *pgIndexPatternRepository) FindAll(ctx context.Context) ([]domain.UtmIndexPattern, error) {
	var items []domain.UtmIndexPattern
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}
