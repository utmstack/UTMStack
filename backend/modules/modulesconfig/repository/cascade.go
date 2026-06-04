package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"
)

type indexPatternToggler struct {
	db *gorm.DB
}

func NewIndexPatternToggler(db *gorm.DB) connectors.IndexPatternToggler {
	return &indexPatternToggler{db: db}
}

func (t *indexPatternToggler) SetActiveByModule(ctx context.Context, moduleName string, active bool) error {
	return t.db.WithContext(ctx).
		Table("utm_index_pattern").
		Where("pattern_module = ?", moduleName).
		Update("is_active", active).Error
}

type logstashFilterToggler struct {
	db *gorm.DB
}

func NewLogstashFilterToggler(db *gorm.DB) connectors.LogstashFilterToggler {
	return &logstashFilterToggler{db: db}
}

func (t *logstashFilterToggler) SetActiveByModule(ctx context.Context, moduleName string, active bool) error {
	now := time.Now().UTC()
	return t.db.WithContext(ctx).
		Table("utm_logstash_filter").
		Where("module_name = ?", moduleName).
		Updates(map[string]any{
			"is_active":  active,
			"updated_at": now,
		}).Error
}

type noopMenuToggler struct{}

func NewNoopMenuToggler() connectors.MenuToggler { return &noopMenuToggler{} }

func (t *noopMenuToggler) SetActiveByModule(_ context.Context, moduleName string, active bool) error {
	logger.Debug("menu toggler (noop): module=" + moduleName + " active=" + boolStr(active))
	return nil
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
