package com.park.utmstack.service.collectors;

import agent.CollectorOuterClass;
import com.park.utmstack.service.application_events.ApplicationEventService;
import com.park.utmstack.service.application_modules.UtmModuleGroupConfigurationService;
import com.park.utmstack.service.dto.collectors.dto.CollectorConfigDTO;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class CollectorService {

    private final CollectorGrpcService collectorGrpcService;
    private final ApplicationEventService eventService;
    private final UtmModuleGroupConfigurationService moduleGroupConfigurationService;
    private final CollectorConfigBuilder CollectorConfigBuilder;

    public void upsertCollectorConfig(CollectorConfigDTO collectorConfig) {

       this.moduleGroupConfigurationService.updateConfigurationKeys(collectorConfig.getModuleId(), collectorConfig.getKeys());

        CollectorOuterClass.CollectorConfig collector = CollectorConfigBuilder.build(collectorConfig);
        collectorGrpcService.upsertCollectorConfig(collector);
    }

}

