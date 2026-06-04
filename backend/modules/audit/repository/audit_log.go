package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/audit/connectors"
	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
	"github.com/utmstack/utmstack/backend/pkg/database"
	"gorm.io/gorm"
)

var auditFilterFields = []string{
	"user_id", "user_login", "action", "status",
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

func (r *pgRepo) Insert(ctx context.Context, log *domain.AuditLog) error {
	return r.db.Create(ctx, log)
}

func (r *pgRepo) List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.AuditLog], error) {
	return r.GetAll(ctx, req, auditFilterFields, "id DESC")
}
