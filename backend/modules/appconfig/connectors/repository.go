package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/appconfig/domain"
)

type Repository interface {
	List(ctx context.Context) ([]domain.Config, error)
	GetByKey(ctx context.Context, key string) (*domain.Config, error)
	GetOwn(ctx context.Context, key string) (*domain.Config, error)
	Save(ctx context.Context, c *domain.Config) error
}
