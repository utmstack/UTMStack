package agent

import (
	"github.com/utmstack/UTMStack/agent-manager/models"
)

// protoToModelCollectorGroupConfiguration Convert a single CollectorGroupConfigurations from proto to model.
func protoToModelCollectorGroupConfiguration(protoConfig *CollectorGroupConfigurations) models.CollectorGroupConfigurations {
	return models.CollectorGroupConfigurations{
		ConfigGroupID:   uint(protoConfig.GroupId),
		ConfKey:         protoConfig.ConfKey,
		ConfValue:       protoConfig.ConfValue,
		ConfName:        protoConfig.ConfName,
		ConfDescription: protoConfig.ConfDescription,
		ConfDataType:    protoConfig.ConfDataType,
		ConfRequired:    protoConfig.ConfRequired,
	}
}

// protoToModelCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from proto to model.
func protoToModelCollectorGroupConfigurations(protoConfigs []*CollectorGroupConfigurations) []models.CollectorGroupConfigurations {
	var modelConfigs []models.CollectorGroupConfigurations
	for _, pc := range protoConfigs {
		modelConfigs = append(modelConfigs, protoToModelCollectorGroupConfiguration(pc))
	}
	return modelConfigs
}

func protoToModelCollectorGroups(protoConfigs []*CollectorConfigGroup, collectorId uint) []models.CollectorConfigGroup {
	var modelConfigs []models.CollectorConfigGroup
	for _, config := range protoConfigs {
		modelConfigs = append(modelConfigs, protoToModelCollectorConfigGroup(config, collectorId))
	}
	return modelConfigs
}

// protoToModelCollectorConfigGroup Convert CollectorConfigGroup from proto to model.
func protoToModelCollectorConfigGroup(protoGroup *CollectorConfigGroup, collectorId uint) models.CollectorConfigGroup {
	return models.CollectorConfigGroup{
		ID:               uint(protoGroup.Id),
		CollectorID:      collectorId,
		GroupName:        protoGroup.GroupName,
		GroupDescription: protoGroup.GroupDescription,
		Configurations:   protoToModelCollectorGroupConfigurations(protoGroup.Configurations),
	}
}

/*
// protoToModelCollector Convert Collector from proto to model.
func protoToModelCollector(proto *Collector) models.Collector {
	id := uint(proto.Id)
	return models.Collector{
		Model: gorm.Model{
			ID: id,
		},
		CollectorKey: proto.CollectorKey,
		Ip:           proto.Ip,
		Hostname:     proto.Hostname,
		Version:      proto.Version,
		Module:       models.CollectorModule(proto.Module.String()),
		Groups:       protoToModelCollectorGroups(proto.Groups, id),
	}
}
*/

// ModelToProtoCollectorGroupConfiguration Convert a single CollectorGroupConfigurations from model to proto.
func modelToProtoCollectorGroupConfiguration(modelConfig models.CollectorGroupConfigurations) *CollectorGroupConfigurations {
	return &CollectorGroupConfigurations{
		GroupId:         int32(modelConfig.ConfigGroupID),
		ConfKey:         modelConfig.ConfKey,
		ConfValue:       modelConfig.ConfValue,
		ConfName:        modelConfig.ConfName,
		ConfDescription: modelConfig.ConfDescription,
		ConfDataType:    modelConfig.ConfDataType,
		ConfRequired:    modelConfig.ConfRequired,
	}
}

// ModelToProtoCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from model to proto.
func modelToProtoCollectorGroupConfigurations(modelConfigs []models.CollectorGroupConfigurations) []*CollectorGroupConfigurations {
	var protoConfigs []*CollectorGroupConfigurations
	for _, mc := range modelConfigs {
		protoConfigs = append(protoConfigs, modelToProtoCollectorGroupConfiguration(mc))
	}
	return protoConfigs
}

// ModelToProtoCollectorGroupConfigurations Convert a slice of CollectorGroupConfigurations from model to proto.
func modelToProtoCollectorGroups(groups []models.CollectorConfigGroup, collectorId int32) []*CollectorConfigGroup {
	var protoGroups []*CollectorConfigGroup
	for _, group := range groups {
		protoGroups = append(protoGroups, modelToProtoCollectorConfigGroup(group, collectorId))
	}
	return protoGroups
}

// ModelToProtoCollectorConfigGroup Convert CollectorConfigGroup from model to proto.
func modelToProtoCollectorConfigGroup(modelGroup models.CollectorConfigGroup, collectorId int32) *CollectorConfigGroup {
	return &CollectorConfigGroup{
		Id:               int32(modelGroup.ID),
		GroupName:        modelGroup.GroupName,
		GroupDescription: modelGroup.GroupDescription,
		Configurations:   modelToProtoCollectorGroupConfigurations(modelGroup.Configurations),
		CollectorId:      collectorId,
	}
}

// ModelToProtoCollector Convert Collector from model to proto.
func modelToProtoCollector(model models.Collector) *Collector {
	collectorStatus, lastSeen := lastSeenService.GetStatus(model.CollectorKey)
	id := int32(model.ID)
	return &Collector{
		Id:           id,
		CollectorKey: model.CollectorKey,
		Ip:           model.Ip,
		Hostname:     model.Hostname,
		Version:      model.Version,
		Status:       Status(collectorStatus),
		LastSeen:     lastSeen,
		Module:       CollectorModule(CollectorModule_value[string(model.Module)]),
		Groups:       modelToProtoCollectorGroups(model.Groups, id),
	}
}
