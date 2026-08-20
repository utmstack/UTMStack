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
	// CountValueContains returns how many rows for `key` (across all tenants)
	// have `needle` as a substring of their JSON value. Used to check whether
	// a branding asset URL is still referenced before deleting its file.
	CountValueContains(ctx context.Context, key, needle string) (int, error)
}
