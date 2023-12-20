package com.park.utmstack.service.dto;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.IntegerFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

public class UtmMenuCriteria implements Serializable {
    /**
     * Class for filtering ModulesNameShort
     */
    public static class ModulesNameShortFilter extends Filter<ModuleName> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private StringFilter name;
    private StringFilter url;
    private LongFilter parentId;
    private IntegerFilter type;
    private LongFilter dashboardId;
    private StringFilter modulesNameShort;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getName() {
        return name;
    }

    public void setName(StringFilter name) {
        this.name = name;
    }

    public StringFilter getUrl() {
        return url;
    }

    public void setUrl(StringFilter url) {
        this.url = url;
    }

    public LongFilter getParentId() {
        return parentId;
    }

    public void setParentId(LongFilter parentId) {
        this.parentId = parentId;
    }

    public IntegerFilter getType() {
        return type;
    }

    public void setType(IntegerFilter type) {
        this.type = type;
    }

    public LongFilter getDashboardId() {
        return dashboardId;
    }

    public void setDashboardId(LongFilter dashboardId) {
        this.dashboardId = dashboardId;
    }

    public StringFilter getModulesNameShort() {
        return modulesNameShort;
    }

    public void setModulesNameShort(StringFilter modulesNameShort) {
        this.modulesNameShort = modulesNameShort;
    }
}
