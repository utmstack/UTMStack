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
public class ModuleBitdefender implements IModule {
    private static final String CLASSNAME = "ModuleBitdefender";

    private final UtmModuleService moduleService;

    public ModuleBitdefender(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.BITDEFENDER);
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

        // connectionKey
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("connectionKey")
            .withConfName("Connection key")
            .withConfDescription("Bitdefender connection key")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());

        // accessUrl
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("accessUrl")
            .withConfName("Access URL")
            .withConfDescription("Bitdefender access URL")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // utmPublicIP
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("utmPublicIP")
            .withConfName("Master public IP or DNS")
            .withConfDescription("Master instance public IP or DNS")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // companyIds
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("companyIds")
            .withConfName("Companies IDs")
            .withConfDescription("Separate the company IDs to be associated with this tenant by commas, for example: BDGZ1234,BDGZ5678,BDGZ9012")
            .withConfDataType("list")
            .withConfRequired(true)
            .build());
        return keys;
    }
}
