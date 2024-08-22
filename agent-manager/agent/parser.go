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
		agentProto := parseAgentToProto(agent)
		agentMessages = append(agentMessages, agentProto)
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
	return &Agent{
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
			Agent:         parseAgentToProto(command.Agent),
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

// func convertToCollectorConfig(collector *models.Collector, configs []*CollectorConfigGroup) *CollectorConfig {
// 	var protoGroups []*CollectorConfigGroup
// 	for _, group := range configs {
// 		intCollectorID, _ := strconv.Atoi(group.CollectorID)
// 		protoGroup := &CollectorConfigGroup{
// 			Id:               int32(group.ID),
// 			GroupName:        group.GroupName,
// 			GroupDescription: group.GroupDescription,
// 			CollectorId:      int32(intCollectorID),
// 		}

// 		for _, config := range group.Configurations {
// 			protoConfig := &CollectorGroupConfigurations{
// 				Id:              int32(config.ID),
// 				GroupId:         int32(config.GroupID),
// 				ConfKey:         config.ConfKey,
// 				ConfValue:       config.ConfValue,
// 				ConfName:        config.ConfName,
// 				ConfDescription: config.ConfDescription,
// 				ConfDataType:    string(config.ConfDataType),
// 				ConfRequired:    config.ConfRequired,
// 			}
// 			protoGroup.Configurations = append(protoGroup.Configurations, protoConfig)
// 		}

// 		protoGroups = append(protoGroups, protoGroup)
// 	}

// 	cnf := &CollectorConfig{
// 		RequestId: "",
// 		Groups:    protoGroups,
// 	}

// 	return cnf
// }

// // protoToModelCollectorGroupConfiguration Convert a single CollectorGroupConfigurations from proto to model.
// func protoToModelCollectorGroupConfiguration(protoConfig *CollectorGroupConfigurations) models.CollectorGroupConfigurations {
// 	return models.CollectorGroupConfigurations{
// 		ConfigGroupID:   uint(protoConfig.GroupId),
// 		ConfKey:         protoConfig.ConfKey,
// 		ConfValue:       protoConfig.ConfValue,
// 		ConfName:        protoConfig.ConfName,
// 		ConfDescription: protoConfig.ConfDescription,
// 		ConfDataType:    protoConfig.ConfDataType,
// 		ConfRequired:    protoConfig.ConfRequired,
// 	}
// }

// // protoToModelCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from proto to model.
// func protoToModelCollectorGroupConfigurations(protoConfigs []*CollectorGroupConfigurations) []models.CollectorGroupConfigurations {
// 	var modelConfigs []models.CollectorGroupConfigurations
// 	for _, pc := range protoConfigs {
// 		modelConfigs = append(modelConfigs, protoToModelCollectorGroupConfiguration(pc))
// 	}
// 	return modelConfigs
// }

// func protoToModelCollectorGroups(protoConfigs []*CollectorConfigGroup, collectorId uint) []models.CollectorConfigGroup {
// 	var modelConfigs []models.CollectorConfigGroup
// 	for _, config := range protoConfigs {
// 		modelConfigs = append(modelConfigs, protoToModelCollectorConfigGroup(config, collectorId))
// 	}
// 	return modelConfigs
// }

// // protoToModelCollectorConfigGroup Convert CollectorConfigGroup from proto to model.
// func protoToModelCollectorConfigGroup(protoGroup *CollectorConfigGroup, collectorId uint) models.CollectorConfigGroup {
// 	return models.CollectorConfigGroup{
// 		ID:               uint(protoGroup.Id),
// 		CollectorID:      collectorId,
// 		GroupName:        protoGroup.GroupName,
// 		GroupDescription: protoGroup.GroupDescription,
// 		Configurations:   protoToModelCollectorGroupConfigurations(protoGroup.Configurations),
// 	}
// }

// /*
// // protoToModelCollector Convert Collector from proto to model.
// func protoToModelCollector(proto *Collector) models.Collector {
// 	id := uint(proto.Id)
// 	return models.Collector{
// 		Model: gorm.Model{
// 			ID: id,
// 		},
// 		CollectorKey: proto.CollectorKey,
// 		Ip:           proto.Ip,
// 		Hostname:     proto.Hostname,
// 		Version:      proto.Version,
// 		Module:       models.CollectorModule(proto.Module.String()),
// 		Groups:       protoToModelCollectorGroups(proto.Groups, id),
// 	}
// }
// */

// // ModelToProtoCollectorGroupConfiguration Convert a single CollectorGroupConfigurations from model to proto.
// func modelToProtoCollectorGroupConfiguration(modelConfig models.CollectorGroupConfigurations) *CollectorGroupConfigurations {
// 	return &CollectorGroupConfigurations{
// 		GroupId:         int32(modelConfig.ConfigGroupID),
// 		ConfKey:         modelConfig.ConfKey,
// 		ConfValue:       modelConfig.ConfValue,
// 		ConfName:        modelConfig.ConfName,
// 		ConfDescription: modelConfig.ConfDescription,
// 		ConfDataType:    modelConfig.ConfDataType,
// 		ConfRequired:    modelConfig.ConfRequired,
// 	}
// }

// // ModelToProtoCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from model to proto.
// func modelToProtoCollectorGroupConfigurations(modelConfigs []models.CollectorGroupConfigurations) []*CollectorGroupConfigurations {
// 	var protoConfigs []*CollectorGroupConfigurations
// 	for _, mc := range modelConfigs {
// 		protoConfigs = append(protoConfigs, modelToProtoCollectorGroupConfiguration(mc))
// 	}
// 	return protoConfigs
// }

// // ModelToProtoCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from model to proto.
// func modelToProtoCollectorGroups(groups []models.CollectorConfigGroup, collectorId int32) []*CollectorConfigGroup {
// 	var protoGroups []*CollectorConfigGroup
// 	for _, group := range groups {
// 		protoGroups = append(protoGroups, modelToProtoCollectorConfigGroup(group, collectorId))
// 	}
// 	return protoGroups
// }

// func modelToProtoCollectorConfigGroup(modelGroup models.CollectorConfigGroup, collectorId int32) *CollectorConfigGroup {
// 	return &CollectorConfigGroup{
// 		Id:               int32(modelGroup.ID),
// 		GroupName:        modelGroup.GroupName,
// 		GroupDescription: modelGroup.GroupDescription,
// 		Configurations:   modelToProtoCollectorGroupConfigurations(modelGroup.Configurations),
// 		CollectorId:      collectorId,
// 	}
// }
