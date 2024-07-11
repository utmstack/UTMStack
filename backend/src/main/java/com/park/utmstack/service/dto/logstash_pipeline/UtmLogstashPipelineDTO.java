package com.park.utmstack.service.dto.logstash_pipeline;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.logstash_pipeline.UtmLogstashPipeline;

import javax.validation.constraints.Size;

public class UtmLogstashPipelineDTO {
    private Long id;

    private String pipelineId;

    @Size(max = 200)
    private String pipelineName;

    private String pipelineStatus;

    private String moduleName;

    private Boolean systemOwner;

    @Size(max = 2000)
    private String pipelineDescription;

    private Boolean pipelineInternal = false;


    public UtmLogstashPipelineDTO(){}
    public UtmLogstashPipelineDTO(Long id, String pipelineId,
                                  String pipelineName,
                                  String pipelineStatus,
                                  String moduleName,
                                  Boolean systemOwner,
                                  String pipelineDescription,
                                  Boolean pipelineInternal) {
        this.id = id;
        this.pipelineId = pipelineId;
        this.pipelineName = pipelineName;
        this.pipelineStatus = pipelineStatus;
        this.moduleName = moduleName;
        this.systemOwner = systemOwner;
        this.pipelineDescription = pipelineDescription;
        this.pipelineInternal = pipelineInternal==null?false:pipelineInternal;
    }
    public UtmLogstashPipelineDTO(UtmLogstashPipeline pipeline) {
        this.id = pipeline.getId();
        this.pipelineId = pipeline.getPipelineId();
        this.pipelineName = pipeline.getPipelineName();
        this.pipelineStatus = pipeline.getPipelineStatus();
        this.moduleName = pipeline.getModuleName();
        this.systemOwner = pipeline.getSystemOwner();
        this.pipelineDescription = pipeline.getPipelineDescription();
        this.pipelineInternal = pipeline.getPipelineInternal()==null?false:pipeline.getPipelineInternal();
    }

    @JsonIgnore
    public UtmLogstashPipeline getPipeline(UtmLogstashPipeline utmLogstashPipeline) {
        if (utmLogstashPipeline==null) utmLogstashPipeline = new UtmLogstashPipeline();
        utmLogstashPipeline.setId(this.getId());
        utmLogstashPipeline.setPipelineName(this.getPipelineName());
        utmLogstashPipeline.setModuleName(this.getModuleName());
        utmLogstashPipeline.setPipelineDescription(this.getPipelineDescription());
        utmLogstashPipeline.setPipelineInternal(this.getPipelineInternal()==null?utmLogstashPipeline.getPipelineInternal():this.getPipelineInternal());
        return utmLogstashPipeline;
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getPipelineId() {
        return pipelineId;
    }

    public void setPipelineId(String pipelineId) {
        this.pipelineId = pipelineId;
    }

    public String getPipelineName() {
        return pipelineName;
    }

    public void setPipelineName(String pipelineName) {
        this.pipelineName = pipelineName;
    }

    public String getPipelineStatus() {
        return pipelineStatus;
    }

    public void setPipelineStatus(String pipelineStatus) {
        this.pipelineStatus = pipelineStatus;
    }

    public String getModuleName() {
        return moduleName;
    }

    public void setModuleName(String moduleName) {
        this.moduleName = moduleName;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public String getPipelineDescription() {
        return pipelineDescription;
    }

    public void setPipelineDescription(String pipelineDescription) {
        this.pipelineDescription = pipelineDescription;
    }

    public Boolean getPipelineInternal() {
        return pipelineInternal;
    }

    public void setPipelineInternal(Boolean pipelineInternal) {
        this.pipelineInternal = pipelineInternal;
    }

}

