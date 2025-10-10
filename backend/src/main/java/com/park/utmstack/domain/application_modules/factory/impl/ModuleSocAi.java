package com.park.utmstack.domain.application_modules.factory.impl;

import com.park.utmstack.domain.application_modules.UtmModule;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.application_modules.factory.IModule;
import com.park.utmstack.domain.application_modules.types.ModuleConfigurationKey;
import com.park.utmstack.domain.application_modules.types.ModuleRequirement;
import com.park.utmstack.domain.application_modules.validators.UtmModuleConfigValidator;
import com.park.utmstack.service.application_modules.UtmModuleService;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.stream.Collectors;

@Component
@RequiredArgsConstructor
public class ModuleSocAi implements IModule {
    private static final String CLASSNAME = "ModuleSocAi";

    private final UtmModuleService moduleService;
    private final UtmModuleConfigValidator utmStackConfigValidator;

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

        keys.add(ModuleConfigurationKey.builder()
                .withGroupId(groupId)
                .withConfKey("utmstack.socai.model")
                .withConfName("Select AI Model")
                .withConfDescription("Choose the AI model that SOC AI will use to analyze alerts.")
                .withConfDataType("select")
                .withConfRequired(true)
                .withConfOptions(
                        "[" +
                                "{\"value\": \"gpt-4\", \"label\": \"GPT-4\"}," +
                                "{\"value\": \"gpt-4-0613\", \"label\": \"GPT-4 (0613)\"}," +
                                "{\"value\": \"gpt-4-32k\", \"label\": \"GPT-4 32K\"}," +
                                "{\"value\": \"gpt-4-32k-0613\", \"label\": \"GPT-4 32K (0613)\"}," +
                                "{\"value\": \"gpt-4-turbo\", \"label\": \"GPT-4 Turbo\"}," +
                                "{\"value\": \"gpt-4o\", \"label\": \"GPT-4 Omni\"}," +
                                "{\"value\": \"gpt-4o-mini\", \"label\": \"GPT-4 Omni Mini\"}," +
                                "{\"value\": \"gpt-4.1\", \"label\": \"GPT-4.1\"}," +
                                "{\"value\": \"gpt-4.1-mini\", \"label\": \"GPT-4.1 Mini\"}," +
                                "{\"value\": \"gpt-4.1-nano\", \"label\": \"GPT-4.1 Nano\"}," +
                                "{\"value\": \"gpt-3.5-turbo\", \"label\": \"GPT-3.5 Turbo\"}," +
                                "{\"value\": \"gpt-3.5-turbo-0613\", \"label\": \"GPT-3.5 Turbo (0613)\"}," +
                                "{\"value\": \"gpt-3.5-turbo-16k\", \"label\": \"GPT-3.5 Turbo 16K\"}," +
                                "{\"value\": \"gpt-3.5-turbo-16k-0613\", \"label\": \"GPT-3.5 Turbo 16K (0613)\"}" +
                                "]"
                )
                .build());

        return keys;
    }

    public boolean validateConfiguration(UtmModule module, List<UtmModuleGroupConfiguration> configuration) throws Exception {
        return utmStackConfigValidator.validate(module, configuration);
    }

    @Override
    public ModuleName getName() {
        return ModuleName.SOC_AI;
    }
}
