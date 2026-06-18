package repository

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/soar/connectors"
	"gorm.io/gorm"
)

type pgResolveFilterRepository struct {
	db *gorm.DB
}

func NewResolveFilterRepository(db *gorm.DB) connectors.ResolveFilterRepository {
	return &pgResolveFilterRepository{db: db}
}

func (r *pgResolveFilterRepository) GetAgentPlatforms(ctx context.Context) ([]string, error) {
	var platforms []string
	err := r.db.WithContext(ctx).
		Raw(`SELECT DISTINCT agent_platform FROM utm_alert_response_rule WHERE agent_platform IS NOT NULL`).
		Scan(&platforms).Error
	if platforms == nil {
		platforms = []string{}
	}
	return platforms, err
}

func (r *pgResolveFilterRepository) GetUsers(ctx context.Context) ([]string, error) {
	var users []string
	err := r.db.WithContext(ctx).Raw(`
		SELECT r1.created_by AS users FROM utm_alert_response_rule r1
		UNION
		SELECT r2.last_modified_by FROM utm_alert_response_rule r2 WHERE r2.last_modified_by IS NOT NULL
	`).Scan(&users).Error
	if users == nil {
		users = []string{}
	}
	return users, err
}
