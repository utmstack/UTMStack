package service

import (
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/repository"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AgentGroupService struct {
	repo *repository.AgentGroupRepository
}

func NewAgentGroupService() *AgentGroupService {
	repo := repository.NewAgentGroupRepository()
	return &AgentGroupService{repo: repo}
}

func (s *AgentGroupService) CreateGroup(agent *models.AgentGroup) error {
	return s.repo.Create(agent)
}

func (s *AgentGroupService) UpdateGroup(agent *models.AgentGroup) error {
	return s.repo.Update(agent)
}

func (s *AgentGroupService) DeleteGroup(id uint) (uint, error) {
	return s.repo.Delete(id)
}

// ListAgentsGroups retrieves a paginated list of agents based on the provided search criteria.
func (s *AgentGroupService) ListAgentsGroups(p util.Pagination, f []util.Filter) ([]models.AgentGroup, int64, error) {
	agents, totalCount, err := s.repo.GetByFilter(p, f)
	if err != nil {
		return nil, 0, status.Errorf(codes.Internal, "failed to retrieve agents: %v", err)
	}
	return agents, totalCount, err
}
