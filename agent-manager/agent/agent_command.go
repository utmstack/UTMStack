package agent

import (
	"context"

	"github.com/utmstack/UTMStack/agent-manager/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Grpc) ListAgentCommands(ctx context.Context, req *ListRequest) (*ListAgentsCommandsResponse, error) {
	page := util.NewPaginator(int(req.PageSize), int(req.PageNumber), req.SortBy)

	filter := util.NewFilter(req.SearchQuery)

	commands, total, err := agentCommandService.ListAgentCommands(page, filter)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch agents: %v", err)
	}

	// Convert the agents into AgentMessages
	var commandMessage []*AgentCommand
	for _, command := range commands {
		commandMessage = append(commandMessage, &AgentCommand{
			AgentId:       uint32(command.AgentID),
			Command:       command.Command,
			CommandStatus: AgentCommandStatus(command.CommandStatus),
			Result:        command.Result,
			ExecutedBy:    command.ExecutedBy,
			CmdId:         command.CmdId,
			CreatedAt:     util.ConvertToTimestamp(command.CreatedAt),
			Agent:         parseAgentToProto(command.Agent),
			Reason:        command.Reason,
			OriginId:      command.OriginId,
			OriginType:    command.OriginType,
		})
	}

	// Return the list of agents as a ListAgentsCommandsResponse
	return &ListAgentsCommandsResponse{
		Rows:  commandMessage,
		Total: int32(total),
	}, nil
}
