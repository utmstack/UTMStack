package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type Repository interface {
	InsertBatch(ctx context.Context, logs []*domain.AuditLog) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time, limit int) (int64, error)
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.AuditLog], error)
	GetByID(ctx context.Context, id uint64) (*domain.AuditLog, error)
}
