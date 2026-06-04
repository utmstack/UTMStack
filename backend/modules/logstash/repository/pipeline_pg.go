package repository

import (
	"context"
	"errors"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"
)

type pgPipelineRepository struct {
	db *gorm.DB
}

func NewPipelineRepository(db *gorm.DB) connectors.PipelineRepository {
	return &pgPipelineRepository{db: db}
}

func (r *pgPipelineRepository) GetByID(ctx context.Context, id int64) (*domain.UtmLogstashPipeline, error) {
	var p domain.UtmLogstashPipeline
	err := r.db.WithContext(ctx).First(&p, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgPipelineRepository) AllActivePipelinesByServer(ctx context.Context) ([]domain.UtmLogstashPipeline, error) {
	// utm_module is not yet ported as a full Go module, but the table exists in the DB.
	// Java JPQL: LEFT JOIN UtmModule as um ON ulp.moduleName = um.moduleName
	// translates to: LEFT JOIN utm_module um ON um.module_name = ulp.module_name
	const query = `
		SELECT ulp.* FROM utm_logstash_pipeline ulp
		LEFT JOIN utm_module um ON um.module_name = ulp.module_name
		WHERE um.module_active IS NULL OR um.module_active = true`

	var pipelines []domain.UtmLogstashPipeline
	err := r.db.WithContext(ctx).Raw(query).Scan(&pipelines).Error
	if err != nil {
		// If the utm_module table doesn't exist yet, fall back to returning all pipelines.
		logger.Warn("pipeline: AllActivePipelinesByServer JOIN failed (utm_module missing?), returning all: " + err.Error())
		if fallbackErr := r.db.WithContext(ctx).Find(&pipelines).Error; fallbackErr != nil {
			return nil, fallbackErr
		}
	}
	return pipelines, nil
}

func (r *pgPipelineRepository) SaveAll(ctx context.Context, pipelines []domain.UtmLogstashPipeline) error {
	if len(pipelines) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Save(&pipelines).Error
}

func (r *pgPipelineRepository) GetNextID(ctx context.Context) (int64, error) {
	var id int64
	err := r.db.WithContext(ctx).Raw("SELECT nextval('utm_logstash_pipeline_id_seq')").Scan(&id).Error
	if err != nil {
		logger.Warn("pipeline: GetNextID: sequence utm_logstash_pipeline_id_seq not found: " + err.Error())
		return 0, err
	}
	return id, nil
}

func (r *pgPipelineRepository) DeletePipeline(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&domain.UtmLogstashPipeline{}, "id = ?", id).Error
}
