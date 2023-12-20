package com.park.utmstack.service.dto.incident_response;

import tech.jhipster.service.filter.*;

import java.io.Serializable;

/**
 * Criteria class for the UtmIncidentAction entity. This class is used in UtmIncidentActionResource to receive all the
 * possible filtering options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /utm-incident-actions?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class UtmIncidentActionCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter actionCommand;
    private StringFilter actionDescription;
    private StringFilter actionParams;
    private IntegerFilter actionType;
    private BooleanFilter actionEditable;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getActionCommand() {
        return actionCommand;
    }

    public void setActionCommand(StringFilter actionCommand) {
        this.actionCommand = actionCommand;
    }

    public StringFilter getActionDescription() {
        return actionDescription;
    }

    public void setActionDescription(StringFilter actionDescription) {
        this.actionDescription = actionDescription;
    }

    public StringFilter getActionParams() {
        return actionParams;
    }

    public void setActionParams(StringFilter actionParams) {
        this.actionParams = actionParams;
    }

    public IntegerFilter getActionType() {
        return actionType;
    }

    public void setActionType(IntegerFilter actionType) {
        this.actionType = actionType;
    }

    public BooleanFilter getActionEditable() {
        return actionEditable;
    }

    public void setActionEditable(BooleanFilter actionEditable) {
        this.actionEditable = actionEditable;
    }
}
