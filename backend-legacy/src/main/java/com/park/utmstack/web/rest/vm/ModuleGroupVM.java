package com.park.utmstack.web.rest.vm;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;

public class ModuleGroupVM {
    @NotBlank
    private String name;
    private String description;
    @NotNull
    private Long moduleId;
    String collector;

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public Long getModuleId() {
        return moduleId;
    }

    public void setModuleId(Long moduleId) {
        this.moduleId = moduleId;
    }

    public String getCollector() {
        return collector;
    }

    public void setCollector(String collector) {
        this.collector = collector;
    }
}
