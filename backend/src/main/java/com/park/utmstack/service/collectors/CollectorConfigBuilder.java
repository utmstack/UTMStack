package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass.CollectorConfig;
import agent.CollectorOuterClass.CollectorConfigGroup;
import agent.CollectorOuterClass.CollectorGroupConfigurations;
import com.park.utmstack.config.Constants;
import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;
import com.park.utmstack.repository.UtmModuleGroupConfigurationRepository;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import com.park.utmstack.service.dto.collectors.dto.CollectorDTO;
import com.park.utmstack.service.application_modules.UtmModuleGroupService;
import com.park.utmstack.util.CipherUtil;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;
import java.util.Objects;

@Component
@RequiredArgsConstructor
public class CollectorConfigBuilder {

    private final UtmModuleGroupService moduleGroupService;
    private final UtmModuleGroupConfigurationRepository configRepo;

    public CollectorConfig build(CollectorConfigDTO dto) {

        List<UtmModuleGroupConfiguration> processed = processPasswords(dto.getKeys());

        return buildCollectorConfig(processed, dto.getCollector());
    }


    private List<UtmModuleGroupConfiguration> processPasswords(List<UtmModuleGroupConfiguration> configs) {

        return configs.stream().map(config -> {

            if (Constants.CONF_TYPE_PASSWORD.equals(config.getConfDataType())) {

                UtmModuleGroupConfiguration original = configRepo.findById(config.getId())
                                                        .orElseThrow(() -> new RuntimeException("Configuration id " + config.getId() + " not found"));

                if (Objects.equals(config.getConfValue(), Constants.MASKED_VALUE)) {
                    config.setConfValue(
                            CipherUtil.decrypt(original.getConfValue(), System.getenv(Constants.ENV_ENCRYPTION_KEY))
                    );
                }
            }

            return config;

        }).toList();
    }


    private CollectorConfig buildCollectorConfig(List<UtmModuleGroupConfiguration> keys, CollectorDTO collectorDTO) {

        List<Long> groupIds = keys.stream()
                .map(UtmModuleGroupConfiguration::getGroupId)
                .distinct()
                .toList();

        List<CollectorConfigGroup> groups = new ArrayList<>();

        for (Long groupId : groupIds) {

            moduleGroupService.findOne(groupId).ifPresent(group -> {

                List<CollectorGroupConfigurations> configs =
                        keys.stream()
                                .filter(k -> k.getGroupId().equals(groupId))
                                .map(this::mapToCollectorGroupConfigurations)
                                .toList();

                groups.add(
                        CollectorConfigGroup.newBuilder()
                                .setGroupName(group.getGroupName())
                                .setGroupDescription(group.getGroupDescription())
                                .addAllConfigurations(configs)
                                .setCollectorId(collectorDTO.getId())
                                .build()
                );
            });
        }

        return CollectorConfig.newBuilder()
                .setCollectorId(String.valueOf(collectorDTO.getId()))
                .setRequestId(String.valueOf(System.currentTimeMillis()))
                .addAllGroups(groups)
                .build();
    }


    private CollectorGroupConfigurations mapToCollectorGroupConfigurations(
            UtmModuleGroupConfiguration c) {

        return CollectorGroupConfigurations.newBuilder()
                .setConfKey(c.getConfKey())
                .setConfName(c.getConfName())
                .setConfDescription(c.getConfDescription())
                .setConfDataType(c.getConfDataType())
                .setConfValue(c.getConfValue())
                .setConfRequired(c.getConfRequired())
                .build();
    }
}

