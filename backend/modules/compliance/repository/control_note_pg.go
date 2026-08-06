package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgNoteRepo struct{ db *gorm.DB }

func NewControlNoteRepository(db *gorm.DB) connectors.ControlNoteRepository {
	return &pgNoteRepo{db: db}
}

func (r *pgNoteRepo) Upsert(ctx context.Context, n *domain.UtmComplianceControlNote) error {
	n.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "tenant_id"}, {Name: "framework_key"}, {Name: "control_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"note", "updated_at"}),
	}).Create(n).Error
}

func (r *pgNoteRepo) Delete(ctx context.Context, frameworkKey, controlID string) error {
	return r.db.WithContext(ctx).
		Where("framework_key = ? AND control_id = ?", frameworkKey, controlID).
		Delete(&domain.UtmComplianceControlNote{}).Error
}

func (r *pgNoteRepo) ListByFramework(ctx context.Context, frameworkKey string) (map[string]string, error) {
	var rows []domain.UtmComplianceControlNote
	if err := r.db.WithContext(ctx).
		Where("framework_key = ?", frameworkKey).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, n := range rows {
		out[n.ControlID] = n.Note
	}
	return out, nil
}
