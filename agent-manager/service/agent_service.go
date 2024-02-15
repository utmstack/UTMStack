package service

import (
	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentService struct {
	repo            *repository.AgentRepository
	lastSeenService *LastSeenService
}

func NewAgentService(lastSeenService *LastSeenService) *AgentService {
	repo := repository.NewAgentRepository()
	return &AgentService{repo: repo, lastSeenService: lastSeenService}
}

func (s *AgentService) Create(agent *models.Agent) error {
	return s.repo.Create(agent)
}

func (s *AgentService) Update(agent *models.Agent) error {
	return s.repo.Update(agent)
}

func (s *AgentService) Delete(key uuid.UUID, deletedBy string) (uint, error) {
	return s.repo.DeleteByKey(key, deletedBy)
}

func (s *AgentService) FindByID(id uint) (*models.Agent, error) {
	return s.repo.GetById(id)
}

func (s *AgentService) FindByToken(token string) (*models.Agent, error) {
	return s.repo.GetByToken(uuid.MustParse(token))
}

func (s *AgentService) FindAll() ([]models.Agent, error) {
	return s.repo.GetAll()
}

func (s *AgentService) UpdateAgentType(id uint, agentType uint) (models.Agent, error) {
	return s.repo.UpdateAgentType(id, agentType)
}

func (s *AgentService) UpdateAgentGroup(id uint, agentGroup uint) (models.Agent, error) {
	return s.repo.UpdateAgentGroup(id, agentGroup)
}

func (s *AgentService) FindByHostname(hostname string) (*models.Agent, error) {
	return s.repo.GetByHostname(hostname)
}

// ListAgents retrieves a paginated list of agents based on the provided search criteria.
func (s *AgentService) ListAgents(p util.Pagination, f []util.Filter) ([]models.Agent, int64, error) {
	agents, totalCount, err := s.repo.GetByFilter(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve agents: %v", err)
	}
	return agents, totalCount, err

}

// ListAgentWithCommands retrieves a paginated list of agents with commands based on the provided search criteria.
func (s *AgentService) ListAgentWithCommands(p util.Pagination, f []util.Filter) ([]models.Agent, int64, error) {
	agents, totalCount, err := s.repo.GetAgentsWithCommands(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve agents: %v", err)
	}
	return agents, totalCount, err
}
