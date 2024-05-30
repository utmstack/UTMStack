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
public class ModuleIbmAs400 implements IModule {
    private static final String CLASSNAME = "ModuleIbmAs400";

    private final UtmModuleService moduleService;

    public ModuleIbmAs400(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.AS_400);
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

        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("collector.as400.user")
                .withConfName("UserName")
                .withConfDescription("The AS400 user's name.")
                .withConfDataType("text")
                .withConfRequired(true)
                .build());
        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("collector.as400.password")
                .withConfName("Password")
                .withConfDescription("The AS400 user's password.")
                .withConfDataType("password")
                .withConfRequired(true)
                .build());
        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("collector.as400.hostname")
                .withConfName("Hostname")
                .withConfDescription("The AS400's hostname or IP address.")
                .withConfDataType("text")
                .withConfRequired(true)
                .build());

        return keys;
    }
}
