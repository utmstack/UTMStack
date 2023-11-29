package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentService struct {
	repo            *repository.AgentRepository
	lastSeenService *AgentLastSeenService
}

func NewAgentService() *AgentService {
	repo := repository.NewAgentRepository()
	lastSeenService := NewAgentLastSeenService()
	lastSeenService.Start()
	return &AgentService{repo: repo, lastSeenService: lastSeenService}
}

func (s *AgentService) Create(agent *models.Agent) error {
	return s.repo.Create(agent)
}

func (s *AgentService) Update(agent *models.Agent) error {
	return s.repo.Update(agent)
}

func (s *AgentService) SetAgentLastSeen(agentKey string) error {
	currentTime := time.Now()
	return s.lastSeenService.Set(agentKey, currentTime)
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

func (s *AgentService) GetAgentLastSeen(agent models.Agent) (models.AgentLastSeen, error) {
	return s.lastSeenService.Get(agent.AgentKey)
}

// ListAgentWithCommands retrieves a paginated list of agents with commands based on the provided search criteria.
func (s *AgentService) ListAgentWithCommands(p util.Pagination, f []util.Filter) ([]models.Agent, int64, error) {
	agents, totalCount, err := s.repo.GetAgentsWithCommands(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve agents: %v", err)
	}
	return agents, totalCount, err
}

func (s *AgentService) GetAgentStatus(agent models.Agent) (models.AgentStatus, string) {
	lastSeen, err := s.GetAgentLastSeen(agent)
	lastPing := lastSeen.LastPing.Format("2006-01-02 15:04:05")
	if err != nil {
		return models.Offline, lastPing
	}
	duration := time.Since(lastSeen.LastPing)
	if duration > time.Minute {
		return models.Offline, lastPing
	}
	return models.Online, lastPing
}
