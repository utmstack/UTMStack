package com.park.utmstack.domain.logstash_pipeline.types;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashInput;

import java.util.List;

public class InputConfiguration {
    Long id;
    Integer pipelineId;
    String inputPrettyName;
    String inputPlugin;
    Boolean inputWithSsl;
    Boolean systemOwner;
    List<InputConfigurationKey> configs;

    public InputConfiguration() {
    }

    @JsonIgnore
    public UtmLogstashInput getUtmLogstashInput() {
        UtmLogstashInput input = new UtmLogstashInput();
        input.setId(id);
        input.setInputPrettyName(inputPrettyName);
        input.setInputPlugin(inputPlugin);
        input.setInputWithSsl(inputWithSsl);
        input.setSystemOwner(systemOwner);

        return input;
    }

    public InputConfiguration(UtmLogstashInput utmLogstashInput){
        this.setId(utmLogstashInput.getId());
        this.setPipelineId(utmLogstashInput.getPipelineId());
        this.setInputPrettyName(utmLogstashInput.getInputPrettyName());
        this.setInputPlugin(utmLogstashInput.getInputPlugin());
        this.setInputWithSsl(utmLogstashInput.getInputWithSsl());
        this.setSystemOwner(utmLogstashInput.getSystemOwner());
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getPipelineId() {
        return pipelineId;
    }

    public void setPipelineId(Integer pipelineId) {
        this.pipelineId = pipelineId;
    }

    public String getInputPrettyName() {
        return inputPrettyName;
    }

    public void setInputPrettyName(String inputPrettyName) {
        this.inputPrettyName = inputPrettyName;
    }

    public String getInputPlugin() {
        return inputPlugin;
    }

    public void setInputPlugin(String inputPlugin) {
        this.inputPlugin = inputPlugin;
    }

    public Boolean getInputWithSsl() {
        return inputWithSsl;
    }

    public void setInputWithSsl(Boolean inputWithSsl) {
        this.inputWithSsl = inputWithSsl;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public List<InputConfigurationKey> getConfigs() {
        return configs;
    }

    public void setConfigs(List<InputConfigurationKey> configs) {
        this.configs = configs;
    }
}
