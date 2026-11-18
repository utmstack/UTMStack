package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/storage/domain"
)

// Retention is a property of the tables, which every tenant shares, so this is
// a platform-level surface rather than a per-tenant one.
type Usecase interface {
	Retentions(ctx context.Context) ([]domain.Retention, error)
	SetRetention(ctx context.Context, r domain.Retention) (domain.Retention, error)

	Usage(ctx context.Context) ([]domain.Usage, error)
	Health(ctx context.Context) (domain.Health, error)

	Tiering(ctx context.Context) (domain.Tiering, error)
	EnableTiering(ctx context.Context, o domain.ColdTarget) (domain.Tiering, error)

	// EnsureLocalCold turns on the local cold mount when nothing is
	// configured yet. A configured cold storage of any kind, or an unusable
	// mount, leaves the configuration untouched and reports why.
	EnsureLocalCold(ctx context.Context) error
}
