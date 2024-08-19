package com.park.utmstack.service.dto.collectors.dto;

import com.park.utmstack.domain.application_modules.UtmModuleGroupConfiguration;

import javax.validation.constraints.NotNull;
import java.util.List;

public class CollectorConfigKeysDTO {
        @NotNull
        CollectorDTO collector;
        @NotNull
        private Long moduleId;
        @NotNull
        private List<UtmModuleGroupConfiguration> keys;

        public Long getModuleId() {
            return moduleId;
        }

        public void setModuleId(Long moduleId) {
            this.moduleId = moduleId;
        }

        public List<UtmModuleGroupConfiguration> getKeys() {
            return keys;
        }

        public void setKeys(List<UtmModuleGroupConfiguration> keys) {
            this.keys = keys;
        }

        public CollectorDTO getCollector() {
            return collector;
        }

        public void setCollector(CollectorDTO collector) {
            this.collector = collector;
        }
}
