package com.park.utmstack.domain.application_modules.factory.impl;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.factory.IModule;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.domain.application_modules.validators.UtmModuleConfigValidator;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.service.application_modules.UtmModuleService;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
public class ModuleCrowdStrike implements IModule {
    private static final String CLASSNAME = "ModuleCrowdStrike";

    private final UtmModuleService moduleService;
    private final UtmModuleConfigValidator utmStackConfigValidator;

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.CROWDSTRIKE);
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
            .withConfKey("crowdStrike.client.id")
            .withConfName("Client ID")
            .withConfDescription("CrowdStrike Client ID")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("crowdStrike.client.secret")
                .withConfName("Client Secret")
                .withConfDescription("CrowdStrike Client Secret")
                .withConfDataType("password")
                .withConfRequired(true)
                .build());

        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("crowdStrike.cloud.region.url")
            .withConfName("Cloud Region URL")
            .withConfDescription("CrowdStrike Cloud Region URL")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());


        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("crowdStrike.app.name")
                .withConfName("App Name")
                .withConfDescription("App Name for CrowdStrike integration")
                .withConfDataType("text")
                .withConfRequired(false)
                .build());

        return keys;

    }

    public boolean validateConfiguration(UtmModule module, List<UtmModuleGroupConfiguration> configuration) {

        return utmStackConfigValidator.validate(module, configuration);
    }

    @Override
    public ModuleName getName() {
        return ModuleName.CROWDSTRIKE;
    }
}
