package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/audit/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type Repository interface {
	Insert(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, req common_models.IListRequest) (common_models.ListResponse[domain.AuditLog], error)
	GetByID(ctx context.Context, id uint64) (*domain.AuditLog, error)
}
