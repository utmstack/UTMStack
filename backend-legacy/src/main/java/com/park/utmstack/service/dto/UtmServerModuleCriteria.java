package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmServerModule entity. This class is used in UtmServerModuleResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-server-modules?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmServerModuleCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter serverId;

    private StringFilter moduleName;

    private BooleanFilter needsRestarts;

    private StringFilter prettyName;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getServerId() {
        return serverId;
    }

    public void setServerId(LongFilter serverId) {
        this.serverId = serverId;
    }

    public StringFilter getModuleName() {
        return moduleName;
    }

    public void setModuleName(StringFilter moduleName) {
        this.moduleName = moduleName;
    }

    public BooleanFilter getNeedsRestarts() {
        return needsRestarts;
    }

    public void setNeedsRestarts(BooleanFilter needsRestarts) {
        this.needsRestarts = needsRestarts;
    }

    public StringFilter getPrettyName() {
        return prettyName;
    }

    public void setPrettyName(StringFilter prettyName) {
        this.prettyName = prettyName;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmServerModuleCriteria that = (UtmServerModuleCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(serverId, that.serverId) &&
            Objects.equals(moduleName, that.moduleName) &&
            Objects.equals(needsRestarts, that.needsRestarts) &&
            Objects.equals(prettyName, that.prettyName);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        serverId,
        moduleName,
        needsRestarts,
        prettyName
        );
    }

    @Override
    public String toString() {
        return "UtmServerModuleCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (serverId != null ? "serverId=" + serverId + ", " : "") +
                (moduleName != null ? "moduleName=" + moduleName + ", " : "") +
                (needsRestarts != null ? "needsRestarts=" + needsRestarts + ", " : "") +
                (prettyName != null ? "prettyName=" + prettyName + ", " : "") +
            "}";
    }

}
