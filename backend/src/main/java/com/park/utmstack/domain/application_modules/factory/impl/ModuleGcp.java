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
public class ModuleGcp implements IModule {
    private static final String CLASSNAME = "ModuleGcp";

    private final UtmModuleService moduleService;

    public ModuleGcp(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.GCP);
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

        // jsonKey
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("jsonKey")
            .withConfName("Json Key")
            .withConfDescription("Configure your GCP json key")
            .withConfDataType("file")
            .withConfRequired(true)
            .build());

        // projectId
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("projectId")
            .withConfName("Project ID")
            .withConfDescription("Configure your GCP project ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // subscription
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("subscription")
            .withConfName("Subscription ID")
            .withConfDescription("Configure your GCP subscription ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // topic
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("topic")
            .withConfName("Topic ID")
            .withConfDescription("Configure your GCP topic ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        return keys;
    }
}
