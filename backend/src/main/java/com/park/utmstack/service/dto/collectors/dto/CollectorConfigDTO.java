package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.web.rest.application_modules.UtmModuleGroupConfigurationResource;

public class CollectorConfigDTO {
    CollectorDTO collector;

    public CollectorDTO getCollector() {
        return collector;
    }

    public void setCollector(CollectorDTO collector) {
        this.collector = collector;
    }

    public UtmModuleGroupConfigurationResource.UpdateConfigurationKeysBody getCollectorConfig() {
        return collectorConfig;
    }

    public void setCollectorConfig(UtmModuleGroupConfigurationResource.UpdateConfigurationKeysBody collectorConfig) {
        this.collectorConfig = collectorConfig;
    }

    UtmModuleGroupConfigurationResource.UpdateConfigurationKeysBody collectorConfig;


}
