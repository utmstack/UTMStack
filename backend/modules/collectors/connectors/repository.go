package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/collectors/domain"
)

type CollectorRepository interface {
	Upsert(ctx context.Context, c *domain.UtmCollector) error
	FindAll(ctx context.Context) ([]*domain.UtmCollector, error)
	MarkOfflineExcept(ctx context.Context, activeIDs []int64) error
	DeleteByID(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*domain.UtmCollector, error)
}
