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
public class ModuleAwsIamUser implements IModule {
    private static final String CLASSNAME = "ModuleAwsIamUser";

    private final UtmModuleService moduleService;

    public ModuleAwsIamUser(UtmModuleService moduleService) {
        this.moduleService = moduleService;
    }

    @Override
    public UtmModule getDetails(Long serverId) throws Exception {
        final String ctx = CLASSNAME + ".getDetails";
        try {
            return moduleService.findByServerIdAndModuleName(serverId, ModuleName.AWS_IAM_USER);
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

        // aws_access_key_id
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("aws_access_key_id")
            .withConfName("Access Key")
            .withConfDescription("Configure Aws Iam User access key")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // aws_default_region
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("aws_default_region")
            .withConfName("Default Region")
            .withConfDescription("Configure Aws Iam User default region")
            .withConfDataType("text")
            .withConfRequired(true)
            .build());

        // aws_secret_access_key
        keys.add(ModuleConfigurationKey.builder()
            .withGroupId(groupId)
            .withConfKey("aws_secret_access_key")
            .withConfName("Secret Key")
            .withConfDescription("Configure Aws Iam User secret kew")
            .withConfDataType("password")
            .withConfRequired(true)
            .build());

        return keys;
    }
}