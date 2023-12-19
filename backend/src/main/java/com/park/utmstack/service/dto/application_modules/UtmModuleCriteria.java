package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import tech.jhipster.service.filter.BooleanFilter;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmModule entity. This class is used in UtmModuleResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-modules?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmModuleCriteria implements Serializable {

    /**
     * Class for filtering ModulesNameShort
     */
    public static class ModulesNameShortFilter extends Filter<ModuleName> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;
    private LongFilter serverId;
    private ModulesNameShortFilter moduleName;
    private BooleanFilter moduleActive;
    private StringFilter moduleCategory;
    private BooleanFilter needsRestart;
    private BooleanFilter liteVersion;
    private BooleanFilter isActivatable;
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

    public ModulesNameShortFilter getModuleName() {
        return moduleName;
    }

    public void setModuleName(ModulesNameShortFilter moduleName) {
        this.moduleName = moduleName;
    }

    public BooleanFilter getModuleActive() {
        return moduleActive;
    }

    public void setModuleActive(BooleanFilter moduleActive) {
        this.moduleActive = moduleActive;
    }

    public StringFilter getModuleCategory() {
        return moduleCategory;
    }

    public void setModuleCategory(StringFilter moduleCategory) {
        this.moduleCategory = moduleCategory;
    }

    public BooleanFilter getNeedsRestart() {
        return needsRestart;
    }

    public void setNeedsRestart(BooleanFilter needsRestart) {
        this.needsRestart = needsRestart;
    }

    public BooleanFilter getLiteVersion() {
        return liteVersion;
    }

    public void setLiteVersion(BooleanFilter liteVersion) {
        this.liteVersion = liteVersion;
    }

    public StringFilter getPrettyName() {
        return prettyName;
    }

    public void setPrettyName(StringFilter prettyName) {
        this.prettyName = prettyName;
    }

    public BooleanFilter getIsActivatable() {
        return isActivatable;
    }

    public void setIsActivatable(BooleanFilter isActivatable) {
        this.isActivatable = isActivatable;
    }
}
