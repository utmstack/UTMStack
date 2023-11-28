package service

import (
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentCommandService struct {
	repo *repository.AgentCommandsRepository
}

func NewAgentCommandService() *AgentCommandService {
	repo := repository.NewAgentCommandsRepository()
	return &AgentCommandService{repo: repo}
}

func (s *AgentCommandService) Create(agent *models.AgentCommand) error {
	return s.repo.Create(agent)
}

func (s *AgentCommandService) Update(agent *models.AgentCommand) error {
	return s.repo.Update(agent)
}

//func (s *AgentCommandService) Delete(id uint) error {
//	return s.repo.Delete(id)
//}

func (s *AgentCommandService) FindByID(id uint) (*models.AgentCommand, error) {
	return s.repo.FindById(id)
}

func (s *AgentCommandService) FindAll() ([]*models.AgentCommand, error) {
	return s.repo.FindAll()
}

func (s *AgentCommandService) GetPendingCommands() ([]models.AgentCommand, error) {
	return s.repo.GetPendingCommandByAgentID()
}
func (s *AgentCommandService) UpdateCommandStatusAndResult(agentID uint, cmdID string, status models.AgentCommandStatus, result string) error {
	return s.repo.UpdateCommandStatusAndResult(agentID, cmdID, status, result)
}

// ListAgentCommands retrieves a paginated list of agents commands based on the provided search criteria.
func (s *AgentCommandService) ListAgentCommands(p util.Pagination, f []util.Filter) ([]models.AgentCommand, int64, error) {
	commands, totalCount, err := s.repo.GetByFilter(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve agents: %v", err)
	}
	return commands, totalCount, err

}
