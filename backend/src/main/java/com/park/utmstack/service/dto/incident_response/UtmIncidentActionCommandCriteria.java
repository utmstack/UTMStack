package com.park.utmstack.service.dto.incident_response;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmIncidentActionCommand entity. This class is used in UtmIncidentActionCommandResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-incident-action-commands?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIncidentActionCommandCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private LongFilter actionId;
    private StringFilter osDistribution;
    private StringFilter osVersion;
    private StringFilter command;

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


    public StringFilter getOsDistribution() {
        return osDistribution;
    }

    public void setOsDistribution(StringFilter osDistribution) {
        this.osDistribution = osDistribution;
    }

    public StringFilter getCommand() {
        return command;
    }

    public void setCommand(StringFilter command) {
        this.command = command;
    }

    public StringFilter getOsVersion() {
        return osVersion;
    }

    public void setOsVersion(StringFilter osVersion) {
        this.osVersion = osVersion;
    }
}
