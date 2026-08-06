package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/loganalyzer/connectors"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/domain"
	"github.com/utmstack/utmstack/backend/modules/loganalyzer/dto"
	"gorm.io/gorm"
)

type pgQueryRepository struct{ db *gorm.DB }

func NewQueryRepository(db *gorm.DB) connectors.QueryRepository {
	return &pgQueryRepository{db: db}
}

func (r *pgQueryRepository) Save(ctx context.Context, q *domain.SavedQuery) error {
	return r.db.WithContext(ctx).Save(q).Error
}

func (r *pgQueryRepository) FindByID(ctx context.Context, id uint64) (*domain.SavedQuery, error) {
	var q domain.SavedQuery
	err := r.db.WithContext(ctx).First(&q, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func (r *pgQueryRepository) List(ctx context.Context, f dto.QueryFilter) ([]domain.SavedQuery, int64, error) {
	q := r.db.WithContext(ctx).Model(&domain.SavedQuery{})
	if f.Name != "" {
		q = q.Where("name ILIKE ?", "%"+f.Name+"%")
	}
	if f.Owner != "" {
		q = q.Where("owner = ?", f.Owner)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []domain.SavedQuery
	if err := q.Order("name ASC").Offset(f.Offset()).Limit(f.Limit()).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *pgQueryRepository) Delete(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Delete(&domain.SavedQuery{}, id).Error
}
