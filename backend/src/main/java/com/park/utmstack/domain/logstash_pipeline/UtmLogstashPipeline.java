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

    @Column(name = "parent_pipeline")
    private Integer parentPipeline;

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

    @Column(name = "events_in")
    private Long eventsIn;

    @Column(name = "events_filtered")
    private Long eventsFiltered;

    @Column(name = "events_out")
    private Long eventsOut;

    @Column(name = "reloads_successes")
    private Long reloadsSuccesses;

    @Column(name = "reloads_failures")
    private Long reloadsFailures;

    @Size(max = 50)
    @Column(name = "reloads_last_failure_timestamp")
    private String reloadsLastFailureTimestamp;

    @Column(name = "reloads_last_error")
    private String reloadsLastError;

    @Size(max = 50)
    @Column(name = "reloads_last_success_timestamp")
    private String reloadsLastSuccessTimestamp;

    public UtmLogstashPipeline(){}
    public UtmLogstashPipeline(Long id, String pipelineId,
                               String pipelineName,
                               Integer parentPipeline,
                               String pipelineStatus,
                               String moduleName,
                               Boolean systemOwner,
                               String pipelineDescription,
                               Boolean pipelineInternal,
                               Long eventsIn,
                               Long eventsFiltered,
                               Long eventsOut,
                               Long reloadsSuccesses,
                               Long reloadsFailures,
                               String reloadsLastFailureTimestamp,
                               String reloadsLastError,
                               String reloadsLastSuccessTimestamp) {
        this.id = id;
        this.pipelineId = pipelineId;
        this.pipelineName = pipelineName;
        this.parentPipeline = parentPipeline;
        this.pipelineStatus = pipelineStatus;
        this.moduleName = moduleName;
        this.systemOwner = systemOwner;
        this.pipelineDescription = pipelineDescription;
        this.pipelineInternal = pipelineInternal==null?false:pipelineInternal;
        this.eventsIn = eventsIn==null?0L:eventsIn;
        this.eventsFiltered = eventsFiltered==null?0L:eventsFiltered;
        this.eventsOut = eventsOut==null?0L:eventsOut;
        this.reloadsSuccesses = reloadsSuccesses==null?0L:reloadsSuccesses;
        this.reloadsFailures = reloadsFailures==null?0L:reloadsFailures;
        this.reloadsLastFailureTimestamp = reloadsLastFailureTimestamp;
        this.reloadsLastError = reloadsLastError;
        this.reloadsLastSuccessTimestamp = reloadsLastSuccessTimestamp;
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

    public Integer getParentPipeline() {
        return parentPipeline;
    }

    public void setParentPipeline(Integer parentPipeline) {
        this.parentPipeline = parentPipeline;
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

    public Long getEventsIn() {
        return eventsIn;
    }

    public void setEventsIn(Long eventsIn) {
        this.eventsIn = eventsIn;
    }

    public Long getEventsFiltered() {
        return eventsFiltered;
    }

    public void setEventsFiltered(Long eventsFiltered) {
        this.eventsFiltered = eventsFiltered;
    }

    public Long getEventsOut() {
        return eventsOut;
    }

    public void setEventsOut(Long eventsOut) {
        this.eventsOut = eventsOut;
    }

    public Long getReloadsSuccesses() {
        return reloadsSuccesses;
    }

    public void setReloadsSuccesses(Long reloadsSuccesses) {
        this.reloadsSuccesses = reloadsSuccesses;
    }

    public Long getReloadsFailures() {
        return reloadsFailures;
    }

    public void setReloadsFailures(Long reloadsFailures) {
        this.reloadsFailures = reloadsFailures;
    }

    public String getReloadsLastFailureTimestamp() {
        return reloadsLastFailureTimestamp;
    }

    public void setReloadsLastFailureTimestamp(String reloadsLastFailureTimestamp) {
        this.reloadsLastFailureTimestamp = reloadsLastFailureTimestamp;
    }

    public String getReloadsLastError() {
        return reloadsLastError;
    }

    public void setReloadsLastError(String reloadsLastError) {
        this.reloadsLastError = reloadsLastError;
    }

    public String getReloadsLastSuccessTimestamp() {
        return reloadsLastSuccessTimestamp;
    }

    public void setReloadsLastSuccessTimestamp(String reloadsLastSuccessTimestamp) {
        this.reloadsLastSuccessTimestamp = reloadsLastSuccessTimestamp;
    }

    public void setDefaults(){
        this.systemOwner = false;
        this.pipelineInternal = this.pipelineInternal==null?false:this.pipelineInternal;
        this.eventsIn = this.eventsIn==null?0L:this.eventsIn;
        this.eventsFiltered = this.eventsFiltered==null?0L:this.eventsFiltered;
        this.eventsOut = this.eventsOut==null?0L:this.eventsOut;
        this.reloadsSuccesses = this.reloadsSuccesses==null?0L:this.reloadsSuccesses;
        this.reloadsFailures = this.reloadsFailures==null?0L:this.reloadsFailures;
        this.pipelineStatus = PipelineStatus.PIPELINE_STATUS_DOWN.get();
    }
}
