package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"gorm.io/gorm"
)

type pgPipelineFilterRepository struct {
	db *gorm.DB
}

func NewPipelineFilterRepository(db *gorm.DB) connectors.PipelineFilterRepository {
	return &pgPipelineFilterRepository{db: db}
}

func (r *pgPipelineFilterRepository) GetFilters(ctx context.Context, pipelineID int32) ([]domain.UtmGroupLogstashPipelineFilters, error) {
	var items []domain.UtmGroupLogstashPipelineFilters
	if err := r.db.WithContext(ctx).Where("pipeline_id = ?", pipelineID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *pgPipelineFilterRepository) Save(ctx context.Context, rel *domain.UtmGroupLogstashPipelineFilters) error {
	return r.db.WithContext(ctx).Create(rel).Error
}

func (r *pgPipelineFilterRepository) DeleteByFilterID(ctx context.Context, filterID int32) error {
	return r.db.WithContext(ctx).
		Where("filter_id = ?", filterID).
		Delete(&domain.UtmGroupLogstashPipelineFilters{}).Error
}

func (r *pgPipelineFilterRepository) DeleteRelationsByPipelineID(ctx context.Context, pipelineID int32) error {
	return r.db.WithContext(ctx).
		Where("pipeline_id = ?", pipelineID).
		Delete(&domain.UtmGroupLogstashPipelineFilters{}).Error
}
