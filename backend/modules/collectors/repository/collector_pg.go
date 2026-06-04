package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/collectors/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type collectorPGRepository struct {
	db *gorm.DB
}

func NewCollectorPGRepository(db *gorm.DB) *collectorPGRepository {
	return &collectorPGRepository{db: db}
}

func (r *collectorPGRepository) Upsert(ctx context.Context, c *domain.UtmCollector) error {
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"status", "collector_key", "ip", "hostname",
				"version", "module", "last_seen", "active",
			}),
		}).
		Create(c)
	return result.Error
}

func (r *collectorPGRepository) FindAll(ctx context.Context) ([]*domain.UtmCollector, error) {
	var collectors []*domain.UtmCollector
	if err := r.db.WithContext(ctx).Find(&collectors).Error; err != nil {
		return nil, fmt.Errorf("collectorPGRepository.FindAll: %w", err)
	}
	return collectors, nil
}

func (r *collectorPGRepository) MarkOfflineExcept(ctx context.Context, activeIDs []int64) error {
	if len(activeIDs) == 0 {
		// Mark everything offline when the response is empty.
		return r.db.WithContext(ctx).
			Model(&domain.UtmCollector{}).
			Updates(map[string]any{"active": false, "status": "OFFLINE"}).Error
	}
	return r.db.WithContext(ctx).
		Model(&domain.UtmCollector{}).
		Where("id NOT IN ?", activeIDs).
		Updates(map[string]any{"active": false, "status": "OFFLINE"}).Error
}

func (r *collectorPGRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmCollector{}, id).Error
}

func (r *collectorPGRepository) FindByID(ctx context.Context, id int64) (*domain.UtmCollector, error) {
	var c domain.UtmCollector
	err := r.db.WithContext(ctx).First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, gorm.ErrRecordNotFound
	}
	return &c, err
}
