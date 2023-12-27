package com.park.utmstack.service.dto;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmConfigurationSection entity. This class is used in UtmConfigurationSectionResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-configuration-sections?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmConfigurationSectionCriteria implements Serializable {

    /**
     * Class for filtering ModulesNameShort
     */
    public static class ModulesNameShortFilter extends Filter<ModuleName> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter section;
    private StringFilter description;
    private ModulesNameShortFilter moduleNameShort;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getSection() {
        return section;
    }

    public void setSection(StringFilter section) {
        this.section = section;
    }

    public StringFilter getDescription() {
        return description;
    }

    public void setDescription(StringFilter description) {
        this.description = description;
    }

    public ModulesNameShortFilter getModuleNameShort() {
        return moduleNameShort;
    }

    public void setModuleNameShort(ModulesNameShortFilter moduleNameShort) {
        this.moduleNameShort = moduleNameShort;
    }
}
