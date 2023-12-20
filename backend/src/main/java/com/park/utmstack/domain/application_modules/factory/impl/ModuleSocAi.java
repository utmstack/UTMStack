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
public class ModuleSocAi implements IModule {
    private static final String CLASSNAME = "ModuleSocAi";

    private final UtmModuleService moduleService;

    public ModuleSocAi(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.SOC_AI);
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

        // soc_ai_key
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("utmstack.socai.key")
            .withConfName("Key")
            .withConfDescription("OpenAI Connection key")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("utmstack.socai.incidentCreation")
            .withConfName("Automatic Incident creation")
            .withConfDescription("If set to \"true\", the system will create incidents based on analysis of alerts.")
            .withConfDataType("bool")
            .withConfRequired(false)
            .build());
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("utmstack.socai.changeAlertStatus")
            .withConfName("Change Alert Status")
            .withConfDescription("If set to \"true\", SOC Ai will automatically change the status of alerts. " +
                "Analysts should investigate those with the status \"In Review\".")
            .withConfDataType("bool")
            .withConfRequired(false)
            .build());

        return keys;
    }
}
