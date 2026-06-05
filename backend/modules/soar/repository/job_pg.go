package repository

import (
	"errors"
	"fmt"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"github.com/utmstack/utmstack/backend/modules/soar/dto"
	"gorm.io/gorm"
)

var _ connectors.JobRepository = (*jobRepository)(nil)

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) *jobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) Save(job *domain.UtmIncidentJob) error {
	if err := r.db.Save(job).Error; err != nil {
		return fmt.Errorf("jobRepository.Save: %w", err)
	}
	return nil
}

func (r *jobRepository) FindByID(id int64) (*domain.UtmIncidentJob, error) {
	var j domain.UtmIncidentJob
	err := r.db.First(&j, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrIncidentRecordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("jobRepository.FindByID: %w", err)
	}
	return &j, nil
}

func (r *jobRepository) applyFilters(q *gorm.DB, f dto.JobFilter) *gorm.DB {
	if f.ActionID != nil {
		q = q.Where("action_id = ?", *f.ActionID)
	}
	if f.Agent != nil {
		q = q.Where("agent = ?", *f.Agent)
	}
	if f.Status != nil {
		q = q.Where("status = ?", *f.Status)
	}
	if f.OriginID != nil {
		q = q.Where("origin_id = ?", *f.OriginID)
	}
	if f.OriginType != nil {
		q = q.Where("origin_type = ?", *f.OriginType)
	}
	return q
}

func (r *jobRepository) FindAll(f dto.JobFilter) ([]domain.UtmIncidentJob, int64, error) {
	q := r.applyFilters(r.db.Model(&domain.UtmIncidentJob{}), f)

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("jobRepository.FindAll count: %w", err)
	}

	page, size := normalizePage(f.Page, f.Size)
	var items []domain.UtmIncidentJob
	if err := q.Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		return nil, 0, fmt.Errorf("jobRepository.FindAll: %w", err)
	}
	return items, total, nil
}

func (r *jobRepository) Count(f dto.JobFilter) (int64, error) {
	var total int64
	q := r.applyFilters(r.db.Model(&domain.UtmIncidentJob{}), f)
	if err := q.Count(&total).Error; err != nil {
		return 0, fmt.Errorf("jobRepository.Count: %w", err)
	}
	return total, nil
}

func (r *jobRepository) Delete(id int64) error {
	if err := r.db.Delete(&domain.UtmIncidentJob{}, id).Error; err != nil {
		return fmt.Errorf("jobRepository.Delete: %w", err)
	}
	return nil
}
