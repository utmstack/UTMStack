package usecase

import (
	"context"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/utmstack/backend/modules/integrations/connectors"
	"github.com/utmstack/utmstack/backend/modules/integrations/domain"
	"gorm.io/gorm"
)

type Bootstrap struct {
	db    *gorm.DB
	store connectors.TenantRepository
}

func NewBootstrap(db *gorm.DB, store connectors.TenantRepository) *Bootstrap {
	return &Bootstrap{db: db, store: store}
}

type legacyConfigRow struct {
	ModuleName   string `gorm:"column:module_name"`
	ModuleActive bool   `gorm:"column:module_active"`
	GroupName    string `gorm:"column:group_name"`
	ConfKey      string `gorm:"column:conf_key"`
	ConfValue    string `gorm:"column:conf_value"`
}

func (b *Bootstrap) Run(ctx context.Context) error {
	if !b.db.Migrator().HasTable("utm_module_group") {
		return nil
	}

	const query = `
		SELECT m.module_name   AS module_name,
		       m.module_active AS module_active,
		       g.group_name    AS group_name,
		       c.conf_key      AS conf_key,
		       c.conf_value    AS conf_value
		FROM utm_module_group g
		JOIN utm_module m ON m.id = g.module_id
		LEFT JOIN utm_module_group_configuration c ON c.group_id = g.id
		ORDER BY m.module_name, g.id`

	var rows []legacyConfigRow
	if err := b.db.WithContext(ctx).Raw(query).Scan(&rows).Error; err != nil {
		return err
	}

	type tkey struct{ module, group string }
	byModule := map[string][]domain.Tenant{}
	active := map[string]bool{}
	pos := map[tkey]int{}

	for _, r := range rows {
		active[r.ModuleName] = r.ModuleActive
		k := tkey{r.ModuleName, r.GroupName}
		i, ok := pos[k]
		if !ok {
			byModule[r.ModuleName] = append(byModule[r.ModuleName], domain.Tenant{
				Name:   r.GroupName,
				Config: map[string]string{},
			})
			i = len(byModule[r.ModuleName]) - 1
			pos[k] = i
		}
		if r.ConfKey != "" {
			byModule[r.ModuleName][i].Config[r.ConfKey] = r.ConfValue
		}
	}

	for module, tenants := range byModule {
		if active[module] {
			if err := b.store.Save(module, tenants); err != nil {
				return err
			}
		} else if err := b.store.SetActiveByModule(ctx, module, false); err != nil {
			return err
		}
	}
	catcher.Info("integrations: migrated legacy module config to tenant files",
		map[string]any{"modules": len(byModule)})

	// Drop the legacy tables (completion marker). Child first (FK).
	if err := b.db.WithContext(ctx).Exec("DROP TABLE IF EXISTS utm_module_group_configuration").Error; err != nil {
		return err
	}
	if err := b.db.WithContext(ctx).Exec("DROP TABLE IF EXISTS utm_module_group").Error; err != nil {
		return err
	}
	return nil
}
