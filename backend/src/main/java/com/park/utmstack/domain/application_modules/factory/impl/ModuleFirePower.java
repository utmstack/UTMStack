package com.park.utmstack.domain.application_modules.factory.impl;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.factory.IModule;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.service.application_modules.UtmModuleService;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;


@Component
public class ModuleFirePower implements IModule {

    private static final String CLASSNAME = "ModuleFirePower";

    private final UtmModuleService moduleService;

    public ModuleFirePower(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.FIRE_POWER);
        } catch (Exception e) {
            throw new Exception(ctx + ": " + e.getMessage());
        }
    }

    @Override
    public List<ModuleRequirement> checkRequirements(Long serverId) throws Exception {
        return Collections.emptyList();
    }

    @Override
    public List<ModuleConfigurationKey> getConfigurationKeys(Long groupId) throws Exception {
        List<ModuleConfigurationKey> keys = new ArrayList<>();

        // eventHubConnection
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("eventHubConnection")
            .withConfName("Event Hub Connection")
            .withConfDescription("Configure the event hub connection")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // consumerGroup
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("consumerGroup")
            .withConfName("Consumer Group")
            .withConfDescription("Configure the consumer group")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // storageContainer
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("storageContainer")
            .withConfName("Storage Container")
            .withConfDescription("Configure the storage container")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // storageConnection
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("storageConnection")
            .withConfName("Storage Connection")
            .withConfDescription("Configure the storage connection")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());
        return keys;
    }

}
