package repository

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/connectors"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gorm.io/gorm"
)

// indexPatternToggler updates utm_index_pattern.is_active. We touch the
// sibling module's table directly (a single column update keyed on
// pattern_module) rather than importing the indexpattern usecase: it keeps
// modulesconfig free of inward dependencies and avoids inflating the
// IndexPatternUsecase interface with a method only the activation cascade
// needs.
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

// logstashFilterToggler updates utm_logstash_filter.is_active and stamps
// updated_at, mirroring the legacy UtmLogstashFilterService.saveAll flow.
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

// noopMenuToggler stands in until a menu module exists in the Go panel. The
// activation cascade calls it but it's intentionally inert; the legacy menu
// cascade hid sidebar entries the new panel doesn't render the same way.
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
