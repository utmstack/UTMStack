package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"gorm.io/gorm"
)

var auditFilterFields = []string{
	"user_id", "user_email", "action", "status",
	"resource_type", "resource_id", "event_type", "ip", "timestamp",
}

type pgRepo struct {
	database.AbstractRepository[domain.AuditLog, uint64]
	db *database.DB
}

func NewRepository(db *gorm.DB) connectors.Repository {
	provider := database.New(db)
	return &pgRepo{
		AbstractRepository: database.NewAbstractRepository[domain.AuditLog, uint64](provider),
		db:                 provider,
	}
}

func (r *pgRepo) InsertBatch(ctx context.Context, logs []*domain.AuditLog) error {
	if len(logs) == 0 {
		return nil
	}
	return r.db.GORM().WithContext(ctx).Create(&logs).Error
}

func (r *pgRepo) DeleteOlderThan(ctx context.Context, cutoff time.Time, limit int) (int64, error) {
	res := r.db.GORM().WithContext(ctx).Exec(`
		DELETE FROM audit_logs
		WHERE id IN (
			SELECT id FROM audit_logs
			WHERE timestamp < ?
			ORDER BY id
			LIMIT ?
			FOR UPDATE SKIP LOCKED
		)`, cutoff, limit)
	return res.RowsAffected, res.Error
}

func (r *pgRepo) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.AuditLog], error) {
	return r.GetAll(ctx, req, auditFilterFields, "id DESC")
}
