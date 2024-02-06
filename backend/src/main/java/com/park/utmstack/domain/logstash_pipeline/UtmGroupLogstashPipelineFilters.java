package com.park.utmstack.domain.logstash_pipeline;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmGroupLogstashPipelineFilters.
 */
@Entity
@Table(name = "utm_group_logstash_pipeline_filters")
public class UtmGroupLogstashPipelineFilters implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id")
    private Long id;

    @Column(name = "filter_id")
    private Integer filterId;

    @Column(name = "pipeline_id")
    private Integer pipelineId;

    @Size(max = 50)
    @Column(name = "relation", length = 50)
    private String relation;

    public UtmGroupLogstashPipelineFilters(){}
    public UtmGroupLogstashPipelineFilters(Long id,
                                           Integer filterId,
                                           Integer pipelineId,
                                           String relation) {
        this.id = id;
        this.filterId = filterId;
        this.pipelineId = pipelineId;
        this.relation = relation;
    }

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Integer getFilterId() {
        return this.filterId;
    }

    public void setFilterId(Integer filterId) {
        this.filterId = filterId;
    }

    public Integer getPipelineId() {
        return this.pipelineId;
    }

    public void setPipelineId(Integer pipelineId) {
        this.pipelineId = pipelineId;
    }

    public String getRelation() {
        return this.relation;
    }

    public void setRelation(String relation) {
        this.relation = relation;
    }
}
