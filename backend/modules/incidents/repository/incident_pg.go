package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"gorm.io/gorm"
)

type pgIncidentRepository struct {
	db *gorm.DB
}

func NewIncidentRepository(db *gorm.DB) connectors.IncidentRepository {
	return &pgIncidentRepository{db: db}
}

func (r *pgIncidentRepository) Save(ctx context.Context, incident *domain.UtmIncident) error {
	return r.db.WithContext(ctx).Create(incident).Error
}

func (r *pgIncidentRepository) Update(ctx context.Context, incident *domain.UtmIncident) error {
	return r.db.WithContext(ctx).Save(incident).Error
}

func (r *pgIncidentRepository) FindByID(ctx context.Context, id int64) (*domain.UtmIncident, error) {
	var row domain.UtmIncident
	if err := r.db.WithContext(ctx).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgIncidentRepository) FindAll(ctx context.Context, q dto.IncidentListQuery) ([]domain.UtmIncident, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := r.db.WithContext(ctx).Model(&domain.UtmIncident{})

	if q.IncidentName != nil && *q.IncidentName != "" {
		db = db.Where("incident_name ILIKE ?", "%"+*q.IncidentName+"%")
	}
	if q.IncidentStatus != nil && *q.IncidentStatus != "" {
		db = db.Where("incident_status = ?", *q.IncidentStatus)
	}
	if q.IncidentAssignedTo != nil && *q.IncidentAssignedTo != "" {
		db = db.Where("incident_assigned_to ILIKE ?", "%"+*q.IncidentAssignedTo+"%")
	}
	if q.CreatedDateStart != nil {
		db = db.Where("incident_created_date >= ?", *q.CreatedDateStart)
	}
	if q.CreatedDateEnd != nil {
		db = db.Where("incident_created_date <= ?", *q.CreatedDateEnd)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	orderBy := "incident_created_date DESC"
	if q.Sort != "" {
		orderBy = parseSortParam(q.Sort)
	}

	var rows []domain.UtmIncident
	if err := db.Order(orderBy).
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *pgIncidentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.UtmIncident{}, id).Error
}

func parseSortParam(sort string) string {
	parts := strings.SplitN(sort, ",", 2)
	if len(parts) == 2 {
		field := strings.TrimSpace(parts[0])
		dir := strings.ToUpper(strings.TrimSpace(parts[1]))
		if dir != "ASC" && dir != "DESC" {
			dir = "ASC"
		}
		return fmt.Sprintf("%s %s", field, dir)
	}
	return sort
}
