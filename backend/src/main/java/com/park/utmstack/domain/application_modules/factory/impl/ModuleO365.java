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
public class ModuleO365 implements IModule {
    private static final String CLASSNAME = "ModuleO365";

    private final UtmModuleService moduleService;

    public ModuleO365(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.O365);
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

        // office365_client_id
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("office365_client_id")
            .withConfName("Client ID")
            .withConfDescription("Configure Office365 client ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // office365_client_secret
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("office365_client_secret")
            .withConfName("Client Secret")
            .withConfDescription("Configure Office365 client secret")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());

        // office365_tenant_id
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("office365_tenant_id")
            .withConfName("Tenant ID")
            .withConfDescription("Configure Office365 tenant ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());
        return keys;
    }
}
