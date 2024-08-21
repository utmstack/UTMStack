package repository

import (
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
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
	err := r.db.Preload("Groups").Preload("Groups.Configurations").First(&collector, id).Error
	if err != nil {
		return nil, err
	}
	return &collector, nil
}

func (r *CollectorRepository) GetCollectorByHostnameAndModule(hostname string, module string) ([]models.Collector, error) {
	var collectors []models.Collector
	err := r.db.Preload("Groups").Preload("Groups.Configurations").Where("hostname = ? and module = ?", hostname, module).Find(&collectors).Error
	if err != nil {
		return nil, err
	}
	return collectors, nil
}

func (r *CollectorRepository) GetCollectorsHostnames() ([]string, error) {
	var hostnames []string
	err := r.db.Model(&models.Collector{}).Distinct("hostname").Pluck("hostname", &hostnames).Error
	if err != nil {
		return nil, err
	}
	return hostnames, nil
}

func (r *CollectorRepository) GetAllCollectors() ([]models.Collector, error) {
	var collectors []models.Collector
	err := r.db.Preload("Groups").Preload("Groups.Configurations").Find(&collectors).Error
	if err != nil {
		return nil, err
	}
	return collectors, nil
}

func (r *CollectorRepository) GetByKey(key uuid.UUID) (*models.Collector, error) {
	var collector models.Collector
	err := r.db.Preload("Groups").Preload("Groups.Configurations").Where("collector_key = ?", key).First(&collector).Error
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

func (r *CollectorRepository) GetCollectorByFilter(p utils.Pagination, f []utils.Filter) ([]models.Collector, int64, error) {
	var collectors []models.Collector
	var count int64
	db := r.db
	tx := db.Model(models.Collector{}).
		Scopes(utils.FilterScope(f)).
		Count(&count).
		Scopes(p.PagingScope).
		Preload("Groups").
		Preload("Groups.Configurations").
		Find(&collectors)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return collectors, count, nil
}

func (r *CollectorRepository) CountCollector() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Collector{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
