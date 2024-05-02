package repository

import (
	"time"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
)

type LastSeenRepository struct {
	db *gorm.DB
}

func GetLastSeenRepository() *LastSeenRepository {
	return &LastSeenRepository{db: config.GetDB()}
}

func (r *LastSeenRepository) GetLastSeen(key string) (models.LastSeen, error) {
	var lastSeen models.LastSeen
	tx := r.db.Where("key = ?", key).First(&lastSeen)
	if tx.Error != nil {
		return models.LastSeen{}, tx.Error
	}
	return lastSeen, nil
}

func (r *LastSeenRepository) ChangeAgentLastSeenByKey(key string) error {
	currentTime := time.Now()
	lastSeen := models.LastSeen{
		Key:      key,
		LastPing: currentTime,
	}
	return r.db.Save(lastSeen).Error
}

func (r *LastSeenRepository) GetAll() ([]models.LastSeen, error) {
	var lastSeen []models.LastSeen
	tx := r.db.Find(&lastSeen)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return lastSeen, nil
}

func (r *LastSeenRepository) Update(lastSeen *models.LastSeen) error {
	return r.db.Save(lastSeen).Error
}
