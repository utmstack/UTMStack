package agent

import (
	"regexp"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
)

func convertModelToAgentResponse(agents []models.Agent, total int64) *ListAgentsResponse {
	var agentMessages []*Agent
	for _, agent := range agents {
		agentMessages = append(agentMessages, parseAgentToProto(agent))
	}
	return &ListAgentsResponse{
		Rows:  agentMessages,
		Total: int32(total),
	}
}

func createHistoryCommand(cmd *UtmCommand, cmdID string, agentId uint) *models.AgentCommand {
	cmdHistory := &models.AgentCommand{
		AgentID:       agentId,
		Command:       cmd.Command,
		CommandStatus: models.Pending,
		Result:        "",
		ExecutedBy:    cmd.ExecutedBy,
		CmdId:         cmdID,
		OriginType:    cmd.OriginType,
		OriginId:      cmd.OriginId,
		Reason:        cmd.Reason,
	}
	return cmdHistory
}

func parseAgentToProto(agent models.Agent) *Agent {
	agentStatus, lastSeen, err := LastSeenServ.GetLastSeenStatus(agent.ID, "agent")
	if err != nil {
		catcher.Error("failed to get last seen status for agent", err, map[string]any{"agent": agent.ID, "process": "agent-manager"})
	}
	agentResult := &Agent{
		Id:             uint32(agent.ID),
		Ip:             agent.Ip,
		Status:         agentStatus,
		Hostname:       agent.Hostname,
		Os:             agent.Os,
		Platform:       agent.Platform,
		Version:        agent.Version,
		AgentKey:       agent.AgentKey,
		LastSeen:       lastSeen,
		Aliases:        agent.Aliases,
		Addresses:      agent.Addresses,
		Mac:            agent.Mac,
		OsMajorVersion: agent.OsMajorVersion,
		OsMinorVersion: agent.OsMinorVersion,
		TenantId:       tenantOrDefault(agent.TenantID),
	}
	return agentResult
}

func convertModelToAgentCommandsProto(commands []models.AgentCommand) []*AgentCommand {
	var commandMessage []*AgentCommand
	for _, command := range commands {
		commandMessage = append(commandMessage, &AgentCommand{
			AgentId:       uint32(command.AgentID),
			Command:       command.Command,
			CommandStatus: AgentCommandStatus(command.CommandStatus),
			Result:        command.Result,
			ExecutedBy:    command.ExecutedBy,
			CmdId:         command.CmdId,
			CreatedAt:     utils.ConvertToTimestamp(command.CreatedAt),
			Reason:        command.Reason,
			OriginId:      command.OriginId,
			OriginType:    command.OriginType,
		})
	}
	return commandMessage
}

func convertModelToCollectorResponse(collectors []models.Collector, total int64) *ListCollectorResponse {
	var collectorMessages []*Collector
	for _, collector := range collectors {
		collectorProto := modelToProtoCollector(collector)
		collectorMessages = append(collectorMessages, collectorProto)
	}
	return &ListCollectorResponse{
		Rows:  collectorMessages,
		Total: int32(total),
	}
}

func replaceSecretValues(input string) string {
	pattern := regexp.MustCompile(`\$\[(\w+):([^]]+)]`)
	return pattern.ReplaceAllStringFunc(input, func(match string) string {
		matches := pattern.FindStringSubmatch(match)
		if len(matches) < 3 {
			return match
		}
		encryptedValue := matches[2]
		decryptedValue, err := utils.DecryptValue(config.EncryptionKey, encryptedValue)
		if err != nil {
			catcher.Error("failed to decrypt secret value in command", err, map[string]any{"process": "agent-manager"})
			return match
		}
		return decryptedValue
	})
}

func modelToProtoCollector(model models.Collector) *Collector {
	collectorStatus, lastSeen, err := LastSeenServ.GetLastSeenStatus(model.ID, "collector")
	if err != nil {
		catcher.Error("failed to get last seen status for collector", err, map[string]any{"model": model.ID, "process": "agent-manager"})
	}
	return &Collector{
		Id:           int32(model.ID),
		CollectorKey: model.CollectorKey,
		Ip:           model.Ip,
		Hostname:     model.Hostname,
		Version:      model.Version,
		Status:       Status(collectorStatus),
		LastSeen:     lastSeen,
		Module:       CollectorModule(CollectorModule_value[string(model.Module)]),
		TenantId:     tenantOrDefault(model.TenantID),
	}
}
