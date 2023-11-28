package repository

import (
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"gorm.io/gorm"
)

type AgentModuleRepository struct {
	db *gorm.DB
}

func NewAgentModulesRepository() *AgentModuleRepository {
	return &AgentModuleRepository{db: config.GetDB()}
}

func (r *AgentModuleRepository) UpdateAgentModule(agentModule []*models.AgentModuleConfiguration) error {
	return r.db.Save(agentModule).Error
}

func (r *AgentModuleRepository) FindByModuleId(id uint) (*models.AgentModule, error) {
	var agentModule models.AgentModule
	err := r.db.First(&agentModule, id).Error
	if err != nil {
		return nil, err
	}
	return &agentModule, nil
}

func (r *AgentModuleRepository) FindAll() ([]*models.AgentModule, error) {
	var agentModules []*models.AgentModule
	err := r.db.Find(&agentModules).Error
	if err != nil {
		return nil, err
	}
	return agentModules, nil
}
