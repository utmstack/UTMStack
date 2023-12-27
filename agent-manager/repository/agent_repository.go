package repository

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"gorm.io/gorm"
)

type AgentRepository struct {
	db *gorm.DB
}

func NewAgentRepository() *AgentRepository {
	return &AgentRepository{db: config.GetDB()}
}

func (r *AgentRepository) Create(agent *models.Agent) error {
	// Begin a transaction
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			// If a panic occurs, rollback the transaction
			tx.Rollback()
		}
	}()
	// Perform the create operation
	if err := tx.Create(agent).Error; err != nil {
		// Rollback the transaction if an error occurs
		return err
	}
	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *AgentRepository) GetById(id uint) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Preload("Commands").First(&agent, id).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetByToken(token uuid.UUID) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Preload("Commands").Where("token = ?", token).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetByHostname(hostname string) (*models.Agent, error) {
	var agent models.Agent
	err := r.db.Where("hostname = ?", hostname).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

func (r *AgentRepository) GetAll() ([]models.Agent, error) {
	var agents []models.Agent
	err := r.db.Preload("Commands").Find(&agents).Error
	if err != nil {
		return nil, err
	}
	return agents, nil
}

func (r *AgentRepository) Update(agent *models.Agent) error {
	return r.db.Save(agent).Error
}

func (r *AgentRepository) DeleteByKey(key uuid.UUID, deletedBy string) (uint, error) {
	var agent models.Agent
	tx := r.db.Where("agent_key = ?", key).First(&agent)
	if tx.Error != nil {
		return 0, tx.Error
	}
	agent.DeletedBy = deletedBy
	r.db.Save(&agent)
	if tx.Error != nil {
		return 0, tx.Error
	}
	tx = r.db.Delete(&agent)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return agent.ID, nil
}

// GetByFilter GetAgentsWithFilters returns a paginated list of agents filtered by search query and sorted by provided fields
func (r *AgentRepository) GetByFilter(p util.Pagination, f []util.Filter) ([]models.Agent, int64, error) {
	var agents []models.Agent
	var count int64
	db := r.db
	tx := db.Model(models.Agent{}).Scopes(util.FilterScope(f)).Count(&count).Scopes(p.PagingScope).Find(&agents)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return agents, count, nil
}

// GetAgentsWithCommands returns a paginated list of agents filtered where at least one command has been executed
func (r *AgentRepository) GetAgentsWithCommands(p util.Pagination, f []util.Filter) ([]models.Agent, int64, error) {
	var agents []models.Agent
	var count int64
	db := r.db
	tx := db.Model(models.Agent{}).
		Scopes(util.FilterScope(f)).
		Count(&count).
		Scopes(p.PagingScope).
		Unscoped().
		Distinct().
		Joins("RIGHT JOIN agent_commands ON agents.id = agent_commands.agent_id").
		Find(&agents)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}
	return agents, count, nil
}

// CountAgents returns the total number of agents
func (r *AgentRepository) CountAgents() (int64, error) {
	var count int64
	if err := r.db.Model(&models.Agent{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *AgentRepository) UpdateAgentType(id uint, agentType uint) (models.Agent, error) {
	var agent models.Agent
	tx := r.db.Where("id = ?", id).Preload("AgentType").Preload("AgentGroup").First(&agent)
	if tx.Error != nil {
		return models.Agent{}, tx.Error
	}
	err := r.db.Model(&models.Agent{}).Where("id = ?", id).Update("agent_type_id", agentType).Error
	if err != nil {
		return models.Agent{}, errors.New(fmt.Sprintf("unable to update agent type with id:%d", id))
	}
	return agent, nil
}

func (r *AgentRepository) UpdateAgentGroup(id uint, agentGroup uint) (models.Agent, error) {
	var agent models.Agent
	tx := r.db.Where("id = ?", id).First(&agent)
	if tx.Error != nil {
		return models.Agent{}, tx.Error
	}
	err := r.db.Model(&models.Agent{}).Where("id = ?", id).Update("agent_group_id", agentGroup).Error
	if err != nil {
		return models.Agent{}, errors.New(fmt.Sprintf("unable to update agent type with id:%d", id))
	}
	return agent, nil
}
