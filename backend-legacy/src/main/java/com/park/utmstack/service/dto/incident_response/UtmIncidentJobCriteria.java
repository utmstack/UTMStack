package com.park.utmstack.service.dto.incident_response;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmIncidentJob entity. This class is used in UtmIncidentJobResource to receive all the possible
 * filtering options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /utm-incident-jobs?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class UtmIncidentJobCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter actionId;

    private StringFilter params;

    private StringFilter agent;

    private IntegerFilter status;

    private StringFilter jobResult;

    private IntegerFilter originId;

    private StringFilter originType;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getActionId() {
        return actionId;
    }

    public void setActionId(LongFilter actionId) {
        this.actionId = actionId;
    }

    public StringFilter getParams() {
        return params;
    }

    public void setParams(StringFilter params) {
        this.params = params;
    }

    public StringFilter getAgent() {
        return agent;
    }

    public void setAgent(StringFilter agent) {
        this.agent = agent;
    }

    public IntegerFilter getStatus() {
        return status;
    }

    public void setStatus(IntegerFilter status) {
        this.status = status;
    }

    public StringFilter getJobResult() {
        return jobResult;
    }

    public void setJobResult(StringFilter jobResult) {
        this.jobResult = jobResult;
    }

    public IntegerFilter getOriginId() {
        return originId;
    }

    public void setOriginId(IntegerFilter originId) {
        this.originId = originId;
    }

    public StringFilter getOriginType() {
        return originType;
    }

    public void setOriginType(StringFilter originType) {
        this.originType = originType;
    }
}
