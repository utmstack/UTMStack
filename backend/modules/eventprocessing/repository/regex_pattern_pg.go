package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/correlation/connectors"
	"github.com/utmstack/utmstack/backend/modules/correlation/domain"
	correrrors "github.com/utmstack/utmstack/backend/modules/correlation/errors"
	"gorm.io/gorm"
)

type pgRegexPatternRepository struct {
	db *gorm.DB
}

func NewRegexPatternRepository(db *gorm.DB) connectors.RegexPatternRepository {
	return &pgRegexPatternRepository{db: db}
}

func (r *pgRegexPatternRepository) Create(ctx context.Context, p *domain.UtmRegexPattern) (*domain.UtmRegexPattern, error) {
	if err := r.db.WithContext(ctx).Create(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (r *pgRegexPatternRepository) Update(ctx context.Context, p *domain.UtmRegexPattern) (*domain.UtmRegexPattern, error) {
	if err := r.db.WithContext(ctx).Save(p).Error; err != nil {
		return nil, err
	}
	return p, nil
}

func (r *pgRegexPatternRepository) GetByID(ctx context.Context, id int64) (*domain.UtmRegexPattern, error) {
	var p domain.UtmRegexPattern
	err := r.db.WithContext(ctx).First(&p, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgRegexPatternRepository) List(ctx context.Context, f connectors.RegexPatternFilters) ([]domain.UtmRegexPattern, int64, error) {
	page, size := normPage(f.Page, f.Size)

	q := r.db.WithContext(ctx).Model(&domain.UtmRegexPattern{})

	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("pattern_id ILIKE ? OR pattern_description ILIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []domain.UtmRegexPattern
	if err := q.Order("id ASC").
		Offset(page * size).
		Limit(size).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgRegexPatternRepository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).Delete(&domain.UtmRegexPattern{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return correrrors.ErrRegexPatternNotFound
	}
	return nil
}

func normPage(page, size int) (int, int) {
	if page < 0 {
		page = 0
	}
	if size < 1 || size > 200 {
		size = 20
	}
	return page, size
}
