package repository

import (
	"time"

	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
)

type AgentLastSeenRepository struct {
	db *gorm.DB
}

func NewAgentLastSeenRepository() *AgentLastSeenRepository {
	return &AgentLastSeenRepository{db: config.GetDB()}
}

func (r *AgentLastSeenRepository) GetLastSeen(agentKey string) (models.AgentLastSeen, error) {
	var lastSeen models.AgentLastSeen
	tx := r.db.Where("agent_key = ?", agentKey).First(&lastSeen)
	if tx.Error != nil {
		return models.AgentLastSeen{}, tx.Error
	}
	return lastSeen, nil
}

func (r *AgentLastSeenRepository) ChangeAgentLastSeenByKey(agentKey string) error {
	currentTime := time.Now()
	lastSeen := models.AgentLastSeen{
		AgentKey: agentKey,
		LastPing: currentTime,
	}
	return r.db.Save(lastSeen).Error
}

func (r *AgentLastSeenRepository) GetAll() ([]models.AgentLastSeen, error) {
	var lastSeen []models.AgentLastSeen
	tx := r.db.Find(&lastSeen)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return lastSeen, nil
}

func (r *AgentLastSeenRepository) Update(lastSeen *models.AgentLastSeen) error {
	return r.db.Save(lastSeen).Error
}
