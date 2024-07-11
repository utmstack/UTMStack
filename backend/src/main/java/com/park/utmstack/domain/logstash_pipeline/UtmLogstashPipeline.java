package com.park.utmstack.domain.logstash_pipeline;

import com.park.utmstack.service.logstash_pipeline.enums.PipelineStatus;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmLogstashPipeline.
 */
@Entity
@Table(name = "utm_logstash_pipeline")
public class UtmLogstashPipeline implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Column(name = "pipeline_id")
    private String pipelineId;

    @Size(max = 200)
    @Column(name = "pipeline_name")
    private String pipelineName;

    @Column(name = "pipeline_status")
    private String pipelineStatus;

    @Column(name = "module_name")
    private String moduleName;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    @Size(max = 2000)
    @Column(name = "pipeline_description")
    private String pipelineDescription;

    @Column(name = "pipeline_internal", columnDefinition = "boolean default false")
    private Boolean pipelineInternal = false;

    @Column(name = "events_out")
    private Long eventsOut;


    public UtmLogstashPipeline(){}
    public UtmLogstashPipeline(Long id, String pipelineId,
                               String pipelineName,
                               String pipelineStatus,
                               String moduleName,
                               Boolean systemOwner,
                               String pipelineDescription,
                               Boolean pipelineInternal,
                               Long eventsOut) {
        this.id = id;
        this.pipelineId = pipelineId;
        this.pipelineName = pipelineName;
        this.pipelineStatus = pipelineStatus;
        this.moduleName = moduleName;
        this.systemOwner = systemOwner;
        this.pipelineDescription = pipelineDescription;
        this.pipelineInternal = pipelineInternal==null?false:pipelineInternal;
        this.eventsOut = eventsOut==null?0L:eventsOut;
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

    public Long getEventsOut() {
        return eventsOut;
    }

    public void setEventsOut(Long eventsOut) {
        this.eventsOut = eventsOut;
    }

    public void setDefaults(){
        this.systemOwner = false;
        this.pipelineInternal = this.pipelineInternal==null?false:this.pipelineInternal;
        this.eventsOut = this.eventsOut==null?0L:this.eventsOut;
        this.pipelineStatus = PipelineStatus.PIPELINE_STATUS_DOWN.get();
    }
}
