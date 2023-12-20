package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmServer entity. This class is used in UtmServerResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-servers?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmServerCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter serverName;

    private StringFilter serverType;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getServerName() {
        return serverName;
    }

    public void setServerName(StringFilter serverName) {
        this.serverName = serverName;
    }

    public StringFilter getServerType() {
        return serverType;
    }

    public void setServerType(StringFilter serverType) {
        this.serverType = serverType;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmServerCriteria that = (UtmServerCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(serverName, that.serverName) &&
            Objects.equals(serverType, that.serverType);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        serverName,
        serverType
        );
    }

    @Override
    public String toString() {
        return "UtmServerCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (serverName != null ? "serverName=" + serverName + ", " : "") +
                (serverType != null ? "serverType=" + serverType + ", " : "") +
            "}";
    }

}
