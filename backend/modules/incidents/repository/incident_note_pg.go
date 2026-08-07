package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/incidents/connectors"
	"github.com/utmstack/utmstack/backend/modules/incidents/domain"
	"github.com/utmstack/utmstack/backend/modules/incidents/dto"
	"gorm.io/gorm"
)

type pgIncidentNoteRepository struct {
	db *gorm.DB
}

func NewIncidentNoteRepository(db *gorm.DB) connectors.IncidentNoteRepository {
	return &pgIncidentNoteRepository{db: db}
}

func (r *pgIncidentNoteRepository) Save(ctx context.Context, note *domain.IncidentNote) error {
	if note.TenantID == uuid.Nil {
		note.TenantID = tenantFromCtx(ctx)
	}
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *pgIncidentNoteRepository) Update(ctx context.Context, note *domain.IncidentNote) error {
	return scopeTenant(ctx, r.db.WithContext(ctx)).Save(note).Error
}

func (r *pgIncidentNoteRepository) FindByIncidentID(ctx context.Context, incidentID uuid.UUID) ([]domain.IncidentNote, error) {
	var rows []domain.IncidentNote
	if err := scopeTenant(ctx, r.db.WithContext(ctx)).Where("incident_id = ?", incidentID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (r *pgIncidentNoteRepository) FindAll(ctx context.Context, q dto.IncidentNoteListQuery) ([]domain.IncidentNote, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.Size < 1 || q.Size > 200 {
		q.Size = 20
	}

	db := scopeTenant(ctx, r.db.WithContext(ctx).Model(&domain.IncidentNote{}))

	if q.IncidentID != nil {
		db = db.Where("incident_id = ?", *q.IncidentID)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []domain.IncidentNote
	if err := db.Order("note_send_date DESC, id ASC").
		Offset((q.Page - 1) * q.Size).
		Limit(q.Size).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
