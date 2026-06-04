package com.park.utmstack.service.dto;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;
import java.util.Objects;

/**
 * Criteria class for the UtmIntegration entity. This class is used in UtmIntegrationResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-integrations?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmIntegrationCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private LongFilter moduleId;

    private StringFilter integrationName;

    private StringFilter integrationDescription;

    private StringFilter url;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public LongFilter getModuleId() {
        return moduleId;
    }

    public void setModuleId(LongFilter moduleId) {
        this.moduleId = moduleId;
    }

    public StringFilter getIntegrationName() {
        return integrationName;
    }

    public void setIntegrationName(StringFilter integrationName) {
        this.integrationName = integrationName;
    }

    public StringFilter getIntegrationDescription() {
        return integrationDescription;
    }

    public void setIntegrationDescription(StringFilter integrationDescription) {
        this.integrationDescription = integrationDescription;
    }

    public StringFilter getUrl() {
        return url;
    }

    public void setUrl(StringFilter url) {
        this.url = url;
    }


    @Override
    public boolean equals(Object o) {
        if (this == o) {
            return true;
        }
        if (o == null || getClass() != o.getClass()) {
            return false;
        }
        final UtmIntegrationCriteria that = (UtmIntegrationCriteria) o;
        return
            Objects.equals(id, that.id) &&
            Objects.equals(moduleId, that.moduleId) &&
            Objects.equals(integrationName, that.integrationName) &&
            Objects.equals(integrationDescription, that.integrationDescription) &&
            Objects.equals(url, that.url);
    }

    @Override
    public int hashCode() {
        return Objects.hash(
        id,
        moduleId,
        integrationName,
        integrationDescription,
        url
        );
    }

    @Override
    public String toString() {
        return "UtmIntegrationCriteria{" +
                (id != null ? "id=" + id + ", " : "") +
                (moduleId != null ? "moduleId=" + moduleId + ", " : "") +
                (integrationName != null ? "integrationName=" + integrationName + ", " : "") +
                (integrationDescription != null ? "integrationDescription=" + integrationDescription + ", " : "") +
                (url != null ? "url=" + url + ", " : "") +
            "}";
    }

}
