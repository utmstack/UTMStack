package com.park.utmstack.domain.logstash_pipeline.types;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInputConfiguration;

public class InputConfigurationKey {
    private Long id;
    private Integer inputId;
    private String confKey;
    private String confValue;
    private String confType;
    private Boolean confRequired;
    private String confValidationRegex;
    private Boolean systemOwner;

    private InputConfigurationKey(){ }

    @JsonIgnore
    public UtmLogstashInputConfiguration getInputConfiguration(Long inputId){
        UtmLogstashInputConfiguration config = new UtmLogstashInputConfiguration();
        config.setId(id);
        config.setInputId(inputId.intValue());
        config.setConfKey(confKey);
        config.setConfType(confType);
        config.setConfValue(confValue);
        config.setConfRequired(confRequired);
        config.setConfValidationRegex(confValidationRegex);
        config.setSystemOwner(systemOwner);

        return config;
    }
    public InputConfigurationKey(UtmLogstashInputConfiguration utmLogstashInputConfiguration){
        this.setId(utmLogstashInputConfiguration.getId());
        this.setInputId(utmLogstashInputConfiguration.getInputId());
        this.setConfKey(utmLogstashInputConfiguration.getConfKey());
        this.setConfType(utmLogstashInputConfiguration.getConfType());
        this.setConfValue(utmLogstashInputConfiguration.getConfValue());
        this.setConfRequired(utmLogstashInputConfiguration.getConfRequired());
        this.setConfValidationRegex(utmLogstashInputConfiguration.getConfValidationRegex());
        this.setSystemOwner(utmLogstashInputConfiguration.getSystemOwner());
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getInputId() {
        return inputId;
    }

    public void setInputId(Integer inputId) {
        this.inputId = inputId;
    }

    public String getConfKey() {
        return confKey;
    }

    public void setConfKey(String confKey) {
        this.confKey = confKey;
    }

    public String getConfValue() {
        return confValue;
    }

    public void setConfValue(String confValue) {
        this.confValue = confValue;
    }

    public String getConfType() {
        return confType;
    }

    public void setConfType(String confType) {
        this.confType = confType;
    }

    public Boolean getConfRequired() {
        return confRequired;
    }

    public void setConfRequired(Boolean confRequired) {
        this.confRequired = confRequired;
    }

    public String getConfValidationRegex() {
        return confValidationRegex;
    }

    public void setConfValidationRegex(String confValidationRegex) {
        this.confValidationRegex = confValidationRegex;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Integer inputId;
        private String confKey;
        private String confValue;
        private String confType;
        private Boolean confRequired;
        private String confValidationRegex;
        private Boolean systemOwner;

        public Builder withInputId(Integer inputId) {
            this.inputId = inputId;
            return this;
        }
        public Builder withConfKey(String confKey) {
            this.confKey = confKey;
            return this;
        }
        public Builder withConfValue(String confValue) {
            this.confValue = confValue;
            return this;
        }
        public Builder withConfType(String confType) {
            this.confType = confType;
            return this;
        }
        public Builder withConfRequired(Boolean confRequired) {
            this.confRequired = confRequired;
            return this;
        }
        public Builder withConfValidationRegex(String confValidationRegex) {
            this.confValidationRegex = confValidationRegex;
            return this;
        }
        public Builder withSystemOwner(Boolean systemOwner) {
            this.systemOwner = systemOwner;
            return this;
        }
        public InputConfigurationKey build(){
            InputConfigurationKey configRow = new InputConfigurationKey();
            configRow.setInputId(inputId);
            configRow.setConfKey(confKey);
            configRow.setConfValue(confValue);
            configRow.setConfType(confType);
            configRow.setConfRequired(confRequired);
            configRow.setConfValidationRegex(confValidationRegex);
            configRow.setSystemOwner(systemOwner);
            return configRow;
        }
    }
}
