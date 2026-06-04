package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"github.com/utmstack/utmstack/backend/modules/soar/domain"
	"gorm.io/gorm"
)

type pgTemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) connectors.TemplateRepository {
	return &pgTemplateRepository{db: db}
}

func (r *pgTemplateRepository) List(ctx context.Context, f connectors.TemplateFilters) ([]domain.AlertResponseActionTemplate, int64, error) {
	page, size := normPage(f.Page, f.Size)

	q := r.db.WithContext(ctx).Model(&domain.AlertResponseActionTemplate{})

	// id.equals
	if f.ID != 0 {
		q = q.Where("id = ?", f.ID)
	}
	// label.contains (maps to the "title" column — Java field is "label", DB column is "title")
	if f.Label != "" {
		q = q.Where("title ILIKE ?", "%"+f.Label+"%")
	}
	// description.contains
	if f.Description != "" {
		q = q.Where("description ILIKE ?", "%"+f.Description+"%")
	}
	// command.contains
	if f.Command != "" {
		q = q.Where("command ILIKE ?", "%"+f.Command+"%")
	}
	// systemOwner.equals
	if f.SystemOwner != nil {
		q = q.Where("system_owner = ?", *f.SystemOwner)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var templates []domain.AlertResponseActionTemplate
	if err := q.Order("id ASC").
		Offset(page * size).
		Limit(size).
		Find(&templates).Error; err != nil {
		return nil, 0, err
	}
	return templates, total, nil
}
