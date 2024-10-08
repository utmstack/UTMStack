package agent

import (
	"regexp"
	"strconv"

	"github.com/utmstack/UTMStack/agent-manager/models"
	"github.com/utmstack/UTMStack/agent-manager/utils"
	"github.com/utmstack/config-client-go/types"
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
		utils.ALogger.ErrorF("failed to get last seen status for agent %d: %v", agent.ID, err)
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
		decryptedValue, _ := utils.DecryptValue(encryptedValue)
		return decryptedValue
	})
}

func modelToProtoCollector(model models.Collector) *Collector {
	collectorStatus, lastSeen, err := LastSeenServ.GetLastSeenStatus(model.ID, "collector")
	if err != nil {
		utils.ALogger.ErrorF("failed to get last seen status for collector %d: %v", model.ID, err)
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
	}
}

func convertModuleGroupToCollectorProto(moduleGroup types.ModuleGroup) *CollectorConfigGroup {
	var protoConfigs []*CollectorGroupConfigurations
	for _, cnf := range moduleGroup.Configurations {
		protoConfigs = append(protoConfigs, &CollectorGroupConfigurations{
			Id:              int32(cnf.ID),
			GroupId:         int32(cnf.GroupID),
			ConfKey:         cnf.ConfKey,
			ConfValue:       cnf.ConfValue,
			ConfName:        cnf.ConfName,
			ConfDescription: cnf.ConfDescription,
			ConfDataType:    string(cnf.ConfDataType),
			ConfRequired:    cnf.ConfRequired,
		})
	}

	intCollectorID, _ := strconv.Atoi(moduleGroup.CollectorID)
	return &CollectorConfigGroup{
		Id:               int32(moduleGroup.ID),
		GroupName:        moduleGroup.GroupName,
		GroupDescription: moduleGroup.GroupDescription,
		CollectorId:      int32(intCollectorID),
		Configurations:   protoConfigs,
	}
}
