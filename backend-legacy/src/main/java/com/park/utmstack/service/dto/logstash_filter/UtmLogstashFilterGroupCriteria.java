package com.park.utmstack.service.dto.logstash_filter;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmLogstashFilterGroup entity. This class is used in UtmLogstashFilterGroupResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-logstash-filter-groups?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmLogstashFilterGroupCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter groupName;

    private StringFilter groupDescription;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getGroupName() {
        return groupName;
    }

    public void setGroupName(StringFilter groupName) {
        this.groupName = groupName;
    }

    public StringFilter getGroupDescription() {
        return groupDescription;
    }

    public void setGroupDescription(StringFilter groupDescription) {
        this.groupDescription = groupDescription;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmLogstashFilterGroupCriteria that = (UtmLogstashFilterGroupCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(groupName, that.groupName) &&
            Objects.equals(groupDescription, that.groupDescription);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        groupName,
        groupDescription
        );
    }

    @Override
    public String toString() {
        return "UtmLogstashFilterGroupCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (groupName != null ? "groupName=" + groupName + ", " : "") +
                (groupDescription != null ? "groupDescription=" + groupDescription + ", " : "") +
            "}";
    }

}
