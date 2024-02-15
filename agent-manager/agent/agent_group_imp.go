package agent

import (
	"context"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) ListGroups(ctx context.Context, req *ListRequest) (*ListAgentsGroupResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	groups, total, err := agentGroupService.ListAgentsGroups(page, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}

	// Convert the agents into AgentMessages
	var groupsMessage []*AgentGroup
	for _, group := range groups {
		groupsMessage = append(groupsMessage, convertGroupAgentToProto(group))
	}

	// Return the list of agents as a ListAgentsCommandsResponse
	return &ListAgentsGroupResponse{
		Rows:  groupsMessage,
		Total: int32(total),
	}, nil
}

func (s *Grpc) CreateGroup(ctx context.Context, req *AgentGroup) (*AgentGroup, error) {
	err := agentGroupService.CreateGroup(&models.AgentGroup{
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch group: %v", err)
	}
	// Return AgentGroup
	return &AgentGroup{
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
	}, nil
}

func (s *Grpc) EditGroup(ctx context.Context, req *AgentGroup) (*AgentGroup, error) {
	err := agentGroupService.UpdateGroup(&models.AgentGroup{
		Id:               uint(req.Id),
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update group: %v", err)
	}
	// Return AgentGroup
	return &AgentGroup{
		GroupName:        req.GroupName,
		GroupDescription: req.GroupDescription,
	}, nil
}

func (s *Grpc) DeleteGroup(ctx context.Context, req *Id) (*Id, error) {
	_, err := agentGroupService.DeleteGroup(uint(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete group: %v", err)
	}
	return req, nil
}

func convertGroupAgentToProto(group models.AgentGroup) *AgentGroup {
	return &AgentGroup{
		Id:               uint32(group.ID),
		GroupName:        group.GroupName,
		GroupDescription: group.GroupDescription,
	}
}
