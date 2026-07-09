package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/compliance/connectors"
	"github.com/utmstack/utmstack/backend/modules/compliance/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgOverrideRepo struct{ db *gorm.DB }

func NewControlStatusOverrideRepository(db *gorm.DB) connectors.ControlStatusOverrideRepository {
	return &pgOverrideRepo{db: db}
}

func (r *pgOverrideRepo) Upsert(ctx context.Context, o *domain.UtmComplianceControlStatusOverride) error {
	o.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "framework_key"}, {Name: "control_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "reason", "updated_at"}),
	}).Create(o).Error
}

func (r *pgOverrideRepo) Delete(ctx context.Context, frameworkKey, controlID string) error {
	return r.db.WithContext(ctx).
		Where("framework_key = ? AND control_id = ?", frameworkKey, controlID).
		Delete(&domain.UtmComplianceControlStatusOverride{}).Error
}

func (r *pgOverrideRepo) ListByFramework(ctx context.Context, frameworkKey string) (map[string]string, error) {
	var rows []domain.UtmComplianceControlStatusOverride
	if err := r.db.WithContext(ctx).
		Where("framework_key = ?", frameworkKey).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, o := range rows {
		out[o.ControlID] = o.Status
	}
	return out, nil
}
