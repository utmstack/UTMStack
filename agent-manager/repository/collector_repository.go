package repository

import (
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"gorm.io/gorm"
)

type CollectorRepository struct {
	db *gorm.DB
}

func NewCollectorRepository() *CollectorRepository {
	return &CollectorRepository{db: config.GetDB()}
}

func (r *CollectorRepository) CreateCollector(collector *models.Collector) error {
	// Begin a transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurs, rollback the transaction
			tx.Rollback()
		}
	}()
	// Perform the create operation
	if err := tx.Create(collector).Error; err != nil {
		// Rollback the transaction if an error occurs
		return err
	}
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *CollectorRepository) GetCollectorById(id uint) (*models.Collector, error) {
	var collector models.Collector
	err := r.db.Preload("Configuration").Preload("Configuration.CollectorGroupConfigs").First(&collector, id).Error
	if err != nil {
		return nil, err
	}
	return &collector, nil
}

func (r *CollectorRepository) GetCollectorByHostname(hostname string) (*models.Collector, error) {
	var collector models.Collector
	err := r.db.Where("hostname = ?", hostname).First(&collector).Error
	if err != nil {
		return nil, err
	}
	return &collector, nil
}

func (r *CollectorRepository) GetAllCollectors() ([]models.Collector, error) {
	var agents []models.Collector
	err := r.db.Preload("Configuration").Preload("Configuration.CollectorGroupConfigs").Find(&agents).Error
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *CollectorRepository) GetByKey(key uuid.UUID) (*models.Collector, error) {
	var collector models.Collector
	err := r.db.Preload("Configuration").Preload("Configuration.CollectorGroupConfigs").Where("token = ?", key).First(&collector).Error
	if err != nil {
		return nil, err
	}
	return &collector, nil
}

func (r *CollectorRepository) UpdateCollector(collector *models.Collector) error {
	return r.db.Save(collector).Error
}

func (r *CollectorRepository) DeleteCollectorByKey(key uuid.UUID, deletedBy string) (uint, error) {
	var collector models.Collector
	tx := r.db.Where("collector_key = ?", key).First(&collector)
	if tx.Error != nil {
		return 0, tx.Error
	}
	collector.DeletedBy = deletedBy
	r.db.Save(&collector)
	if tx.Error != nil {
		return 0, tx.Error
	}
	tx = r.db.Delete(&collector)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return collector.ID, nil
}

// GetCollectorByFilter returns a paginated list of agents filtered by search query and sorted by provided fields
func (r *CollectorRepository) GetCollectorByFilter(p util.Pagination, f []util.Filter) ([]models.Collector, int64, error) {
	var collectors []models.Collector
	var count int64
	db := r.db
	tx := db.Model(models.Collector{}).
		Scopes(util.FilterScope(f)).
		Count(&count).
		Scopes(p.PagingScope).
		Preload("Configuration").
		Preload("Configuration.CollectorGroupConfigs").
		Find(&collectors)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return collectors, count, nil
}

// CountCollector returns the total number of agents
func (r *CollectorRepository) CountCollector() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Collector{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
