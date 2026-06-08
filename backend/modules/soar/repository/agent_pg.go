package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"gorm.io/gorm"
)

// Reaches across into the datasources table (owned by the datasources module)
// because SOAR rule targeting needs OS-platform filtering, and the platform lives
// in the jsonb `metadata` column — which the generic /api/datasources list query
// doesn't filter on. Same cross-module-aggregation pattern as resolve_filter_pg.go.
type pgAgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository(db *gorm.DB) connectors.AgentRepository {
	return &pgAgentRepository{db: db}
}

func (r *pgAgentRepository) ListNamesByPlatform(ctx context.Context, platform string) ([]string, error) {
	var names []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT asset_name FROM datasources
		WHERE source_kind = 'agent'
		  AND asset_name IS NOT NULL
		  AND metadata->>'osPlatform' = ?
		ORDER BY asset_name
	`, platform).Scan(&names).Error
	if names == nil {
		names = []string{}
	}
	return names, err
}
