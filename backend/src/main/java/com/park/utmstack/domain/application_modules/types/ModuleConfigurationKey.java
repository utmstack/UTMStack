package com.park.utmstack.domain.application_modules.types;

public class ModuleConfigurationKey {
    private Long groupId;
    private String confKey;
    private String confName;
    private String confDescription;
    private String confDataType;
    private Boolean confRequired;

    private ModuleConfigurationKey() {
    }

    public Long getGroupId() {
        return groupId;
    }

    public void setGroupId(Long groupId) {
        this.groupId = groupId;
    }

    public String getConfKey() {
        return confKey;
    }

    public void setConfKey(String confKey) {
        this.confKey = confKey;
    }

    public String getConfName() {
        return confName;
    }

    public void setConfName(String confName) {
        this.confName = confName;
    }

    public String getConfDescription() {
        return confDescription;
    }

    public void setConfDescription(String confDescription) {
        this.confDescription = confDescription;
    }

    public String getConfDataType() {
        return confDataType;
    }

    public void setConfDataType(String confDataType) {
        this.confDataType = confDataType;
    }

    public Boolean getConfRequired() {
        return confRequired;
    }

    public void setConfRequired(Boolean confRequired) {
        this.confRequired = confRequired;
    }

    public static Builder builder() {
        return new Builder();
    }

    public static class Builder {
        private Long groupId;
        private String confKey;
        private String confName;
        private String confDescription;
        private String confDataType;
        private Boolean confRequired;

        public Builder withGroupId(Long groupId) {
            this.groupId = groupId;
            return this;
        }

        public Builder withConfKey(String confKey) {
            this.confKey = confKey;
            return this;
        }

        public Builder withConfName(String confName) {
            this.confName = confName;
            return this;
        }

        public Builder withConfDescription(String confDescription) {
            this.confDescription = confDescription;
            return this;
        }

        public Builder withConfDataType(String confDataType) {
            this.confDataType = confDataType;
            return this;
        }

        public Builder withConfRequired(Boolean confRequired) {
            this.confRequired = confRequired;
            return this;
        }

        public ModuleConfigurationKey build() {
            ModuleConfigurationKey key = new ModuleConfigurationKey();
            key.setGroupId(groupId);
            key.setConfKey(confKey);
            key.setConfName(confName);
            key.setConfDescription(confDescription);
            key.setConfDataType(confDataType);
            key.setConfRequired(confRequired);
            return key;
        }
    }
}
