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
public class ModuleSophos implements IModule {
    private static final String CLASSNAME = "ModuleAwsIamUser";

    private final UtmModuleService moduleService;

    public ModuleSophos(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.SOPHOS);
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

        // sophos_api_url
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("sophos_api_url")
            .withConfName("API Url")
            .withConfDescription("Configure Sophos Central api url")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // sophos_authorization
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("sophos_authorization")
            .withConfName("Authorization")
            .withConfDescription("Configure Sophos Central Authorization")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());

        // sophos_x_api_key
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("sophos_x_api_key")
            .withConfName("X-API-KEY")
            .withConfDescription("Configure Sophos Central api key")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());

        return keys;
    }
}
