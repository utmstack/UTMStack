package usecase

// TODO(deploy): ensure the filters directory is mounted in production (env: LOGSTASH_FILTERS_DIR)

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/utmstack/utmstack/backend/modules/logstash/connectors"
	"github.com/utmstack/utmstack/backend/modules/logstash/domain"
	"github.com/utmstack/utmstack/backend/pkg/logger"
	"gopkg.in/yaml.v3"
)

const defaultFiltersDir = "./utmstack/filters"

type FilterYAML struct {
	FilterName     string  `yaml:"filterName"`
	LogstashFilter string  `yaml:"logstashFilter"`
	FilterGroupId  *int64  `yaml:"filterGroupId"`
	DataTypeId     *int64  `yaml:"dataTypeId"`
	ModuleName     *string `yaml:"moduleName"`
	FilterVersion  *string `yaml:"filterVersion"`
	IsActive       bool    `yaml:"isActive"`
}

type FilterDefinitionSyncService struct {
	filterRepo connectors.FilterRepository
	filtersDir string // env: LOGSTASH_FILTERS_DIR, default "./utmstack/filters"
}

func NewFilterDefinitionSyncService(repo connectors.FilterRepository) *FilterDefinitionSyncService {
	dir := os.Getenv("LOGSTASH_FILTERS_DIR")
	if dir == "" {
		dir = defaultFiltersDir
	}
	return &FilterDefinitionSyncService{
		filterRepo: repo,
		filtersDir: dir,
	}
}

func (s *FilterDefinitionSyncService) Sync(ctx context.Context) {
	logger.Info("FilterDefinitionSyncService: starting filter sync from " + s.filtersDir)

	info, err := os.Stat(s.filtersDir)
	if err != nil || !info.IsDir() {
		logger.Warn("FilterDefinitionSyncService: filters directory not found or not a directory: " + s.filtersDir + " — skipping sync")
		return
	}

	walkErr := filepath.Walk(s.filtersDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			logger.Error("FilterDefinitionSyncService: error accessing path " + path + ": " + err.Error())
			return nil // continue walking
		}
		if fi.IsDir() {
			return nil // descend into subdirectories
		}
		if !isYAMLFile(fi.Name()) {
			return nil
		}

		filters, loadErr := loadFiltersFromFile(path)
		if loadErr != nil {
			logger.Error("FilterDefinitionSyncService: failed to load file " + path + ": " + loadErr.Error())
			return nil
		}

		for _, fy := range filters {
			if upsertErr := s.upsertFilter(ctx, fy); upsertErr != nil {
				logger.Error("FilterDefinitionSyncService: failed to upsert filter from file " + path + ": " + upsertErr.Error())
			}
		}
		return nil
	})

	if walkErr != nil {
		logger.Error("FilterDefinitionSyncService: directory walk failed: " + walkErr.Error())
		return
	}

	logger.Info("FilterDefinitionSyncService: filter sync completed")
}

func (s *FilterDefinitionSyncService) upsertFilter(ctx context.Context, fy FilterYAML) error {
	now := time.Now().UTC()

	// Build the domain entity from the YAML definition.
	filter := &domain.UtmLogstashFilter{
		FilterName:     fy.FilterName,
		LogstashFilter: fy.LogstashFilter,
		FilterGroupID:  fy.FilterGroupId,
		DataTypeID:     fy.DataTypeId,
		ModuleName:     fy.ModuleName,
		FilterVersion:  fy.FilterVersion,
		IsActive:       fy.IsActive,
		SystemOwner:    true,
		UpdatedAt:      &now,
	}

	// 1. Try to find by content match (system-owned).
	existing, err := s.filterRepo.FindByLogstashFilterAndSystemOwner(ctx, fy.LogstashFilter)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update — preserve the existing ID.
		filter.ID = existing.ID
		return s.filterRepo.Update(ctx, filter)
	}

	// 2. Fall back to name match.
	if fy.FilterName != "" {
		byName, nameErr := s.filterRepo.FindByFilterName(ctx, fy.FilterName)
		if nameErr != nil {
			return nameErr
		}
		if byName != nil {
			filter.ID = byName.ID
			return s.filterRepo.Update(ctx, filter)
		}
	}

	// 3. Create new system-owned filter.
	return s.filterRepo.Create(ctx, filter)
}

// ── file loaders ───────────────────────────────────────────────────────────────

func isYAMLFile(name string) bool {
	ext := filepath.Ext(name)
	return ext == ".yaml" || ext == ".yml"
}

func loadFiltersFromFile(path string) ([]FilterYAML, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Try as list first.
	var list []FilterYAML
	if err := yaml.Unmarshal(data, &list); err == nil && len(list) > 0 {
		return list, nil
	}

	// Try as single map.
	var single FilterYAML
	if err := yaml.Unmarshal(data, &single); err != nil {
		return nil, err
	}
	return []FilterYAML{single}, nil
}
