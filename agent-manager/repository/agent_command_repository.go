package repository

import (
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"gorm.io/gorm"
)

type AgentCommandsRepository struct {
	db *gorm.DB
}

func NewAgentCommandsRepository() *AgentCommandsRepository {
	return &AgentCommandsRepository{db: config.GetDB()}
}

func (r *AgentCommandsRepository) Create(agentCommandHistory *models.AgentCommand) error {
	return r.db.Create(agentCommandHistory).Error
}

func (r *AgentCommandsRepository) Update(agentCommandHistory *models.AgentCommand) error {
	return r.db.Save(agentCommandHistory).Error
}

func (r *AgentCommandsRepository) Delete(id uint) error {
	return r.db.Delete(&models.AgentCommand{}, id).Error
}

func (r *AgentCommandsRepository) FindById(id uint) (*models.AgentCommand, error) {
	var agentCommandHistory models.AgentCommand
	err := r.db.First(&agentCommandHistory, id).Error
	if err != nil {
		return nil, err
	}
	return &agentCommandHistory, nil
}

// GetByFilter GetAgentsWithFilters returns a paginated list of agents filtered by search query and sorted by provided fields
func (r *AgentCommandsRepository) GetByFilter(p util.Pagination, f []util.Filter) ([]models.AgentCommand, int64, error) {
	var commands []models.AgentCommand
	var count int64
	db := r.db
	tx := db.Model(models.AgentCommand{}).Scopes(util.FilterScope(f)).Count(&count).Scopes(p.PagingScope).Preload("Agent").Unscoped().Find(&commands)
	if tx.Error != nil {
		println(tx.Error)
		return nil, 0, tx.Error
	}

	return commands, count, nil
}

func (r *AgentCommandsRepository) FindAll() ([]*models.AgentCommand, error) {
	var agentCommandHistories []*models.AgentCommand
	err := r.db.Find(&agentCommandHistories).Error
	if err != nil {
		return nil, err
	}
	return agentCommandHistories, nil
}

func (r *AgentCommandsRepository) GetPendingCommandByAgentID() ([]models.AgentCommand, error) {
	var commands []models.AgentCommand
	err := r.db.Where("command_status = ?", models.Queue).First(&commands).Error
	if err != nil {
		return nil, err
	}
	for _, command := range commands {
		command.CommandStatus = models.Pending
		if err := r.db.Save(command).Error; err != nil {
			return nil, err
		}
	}
	return commands, nil
}

func (r *AgentCommandsRepository) UpdateCommandStatusAndResult(agentID uint, cmdID string, status models.AgentCommandStatus, result string) error {
	return r.db.Model(&models.AgentCommand{}).
		Where("agent_id = ? AND cmd_id = ?", agentID, cmdID).
		Updates(map[string]interface{}{
			"command_status": status,
			"result":         result,
		}).Error
}
