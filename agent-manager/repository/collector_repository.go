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

	for i, group := range collector.Groups {
		for j, config := range group.Configurations {
			if config.ConfDataType == "password" {
				decrypted, err := util.DecryptValue(config.ConfValue)
				if err != nil {
					return nil, err
				}
				collector.Groups[i].Configurations[j].ConfValue = decrypted
			}
		}
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

	err := r.DeleteConfigGroupByCollectorID(collector.ID)
	if err != nil {
		return 0, err
	}

	return collector.ID, nil
}

func (r *CollectorRepository) DeleteConfigGroupByCollectorID(collectorID uint) error {
	var configGroups []models.CollectorConfigGroup
	if err := r.db.Where("collector_id = ?", collectorID).Find(&configGroups).Error; err != nil {
		return err
	}
	for _, group := range configGroups {
		if err := r.DeleteConfigsByGroupID(group.ID); err != nil {
			return err
		}
	}
	return r.db.Where("collector_id = ?", collectorID).Delete(&models.CollectorConfigGroup{}).Error
}

func (r *CollectorRepository) DeleteConfigsByGroupID(groupID uint) error {
	return r.db.Where("config_group_id = ?", groupID).Delete(&models.CollectorGroupConfigurations{}).Error
}

// GetCollectorByFilter returns a paginated list of collectors filtered by search query and sorted by provided fields
func (r *CollectorRepository) GetCollectorByFilter(p util.Pagination, f []util.Filter) ([]models.Collector, int64, error) {
	var collectors []models.Collector
	var count int64
	db := r.db
	tx := db.Model(models.Collector{}).
		Scopes(util.FilterScope(f)).
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

// CountCollector returns the total number of collectors
func (r *CollectorRepository) CountCollector() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Collector{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *CollectorRepository) UpdateCollectorConfigGroup(collectorConfigGroup *models.CollectorConfigGroup) error {
	return r.db.Save(collectorConfigGroup).Error
}

func (r *CollectorRepository) UpdateCollectorConfig(groups []models.CollectorConfigGroup, collectorId uint) error {
	for _, group := range groups {
		group.CollectorID = collectorId
		if err := r.db.Save(&group).Error; err != nil {
			return err
		}
		for _, groupConfig := range group.Configurations {
			groupConfig.ConfigGroupID = group.ID
			if groupConfig.ConfDataType == "password" {
				secret, _ := util.EncryptValue(groupConfig.ConfValue)
				groupConfig.ConfValue = secret
			}
			if err := r.db.Save(&groupConfig).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
