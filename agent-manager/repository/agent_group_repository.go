package repository

import (
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"gorm.io/gorm"
)

type AgentGroupRepository struct {
	db *gorm.DB
}

func NewAgentGroupRepository() *AgentGroupRepository {
	return &AgentGroupRepository{db: config.GetDB()}
}

func (r *AgentGroupRepository) Create(agentGroup *models.AgentGroup) error {
	return r.db.Create(agentGroup).Error
}

func (r *AgentGroupRepository) Update(agentGroup *models.AgentGroup) error {
	return r.db.Save(agentGroup).Error
}

func (r *AgentGroupRepository) Delete(id uint) (uint, error) {
	return id, r.db.Delete(&models.AgentGroup{}, id).Error
}

func (r *AgentGroupRepository) FindById(id uint) (*models.AgentGroup, error) {
	var agentGroup models.AgentGroup
	err := r.db.First(&agentGroup, id).Error
	if err != nil {
		return nil, err
	}
	return &agentGroup, nil
}

// GetByFilter returns a paginated list of agents filtered by search query and sorted by provided fields
func (r *AgentGroupRepository) GetByFilter(p util.Pagination, f []util.Filter) ([]models.AgentGroup, int64, error) {
	var groups []models.AgentGroup
	var count int64
	db := r.db
	tx := db.Model(models.AgentGroup{}).Scopes(util.FilterScope(f)).Count(&count).Scopes(p.PagingScope).Find(&groups)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return groups, count, nil
}
