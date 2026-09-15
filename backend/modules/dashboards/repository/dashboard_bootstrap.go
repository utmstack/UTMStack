package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/dashboards/domain"
	"github.com/utmstack/utmstack/backend/pkg/authz"
)

type DashboardBootstrap struct {
	srcDir string
	db     *gorm.DB
}

func NewDashboardBootstrap(srcDir string, db *gorm.DB) *DashboardBootstrap {
	return &DashboardBootstrap{srcDir: srcDir, db: db}
}

func (b *DashboardBootstrap) Run(ctx context.Context) error {
	if _, err := os.Stat(b.srcDir); os.IsNotExist(err) {
		return nil
	}

	ctx = authz.WithTenantID(ctx, authz.DefaultTenantID)

	return filepath.WalkDir(b.srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != DashboardFileExt {
			return nil
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			_ = catcher.Error("dashboards: reading a shipped dashboard definition failed", readErr,
				map[string]any{"file": path})
			return nil
		}

		var def DashboardDefinition
		if yamlErr := yaml.Unmarshal(data, &def); yamlErr != nil {
			_ = catcher.Error("dashboards: parsing a shipped dashboard definition failed", yamlErr,
				map[string]any{"file": path})
			return nil
		}

		if seedErr := b.seedOne(ctx, def); seedErr != nil {
			_ = catcher.Error("dashboards: seeding a system dashboard failed", seedErr,
				map[string]any{"file": path, "name": def.Name})
		}
		return nil
	})
}

func (b *DashboardBootstrap) seedOne(ctx context.Context, def DashboardDefinition) error {
	if def.Name == "" {
		return errors.New("dashboard definition has no name")
	}

	return b.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		var existing domain.Dashboard
		err := tx.Where("name = ?", def.Name).First(&existing).Error
		switch {
		case err == nil:
			existing.Description = def.Description
			existing.SystemOwner = true
			existing.ModifiedDate = now
			if saveErr := tx.Save(&existing).Error; saveErr != nil {
				return fmt.Errorf("updating dashboard row: %w", saveErr)
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			existing = domain.Dashboard{
				Name:         def.Name,
				Description:  def.Description,
				SystemOwner:  true,
				CreatedDate:  now,
				ModifiedDate: now,
			}
			if createErr := tx.Create(&existing).Error; createErr != nil {
				return fmt.Errorf("creating dashboard row: %w", createErr)
			}
		default:
			return fmt.Errorf("looking up dashboard row: %w", err)
		}

		if delErr := tx.Where("dashboard_id = ? AND system_owner = ?", existing.ID, true).
			Delete(&domain.Visualization{}).Error; delErr != nil {
			return fmt.Errorf("clearing old widgets: %w", delErr)
		}

		for i, w := range def.Widgets {
			specJSON, specErr := encodeSpec(w.Spec)
			if specErr != nil {
				return fmt.Errorf("widget %d (%s): %w", i, def.Name, specErr)
			}
			configJSON, err := json.Marshal(w.Config)
			if err != nil {
				return fmt.Errorf("widget %d (%s): encoding config: %w", i, def.Name, err)
			}
			layoutJSON, err := json.Marshal(w.Layout)
			if err != nil {
				return fmt.Errorf("widget %d (%s): encoding layout: %w", i, def.Name, err)
			}

			viz := domain.Visualization{
				DashboardID:  existing.ID,
				Spec:         string(specJSON),
				Config:       string(configJSON),
				Layout:       string(layoutJSON),
				SystemOwner:  true,
				CreatedDate:  now,
				ModifiedDate: now,
			}
			if createErr := tx.Create(&viz).Error; createErr != nil {
				return fmt.Errorf("widget %d (%s): %w", i, def.Name, createErr)
			}
		}
		return nil
	})
}

func encodeSpec(raw map[string]any) (json.RawMessage, error) {
	data, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var spec domain.Spec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, err
	}
	if err := spec.Validate(); err != nil {
		return nil, err
	}
	return data, nil
}
