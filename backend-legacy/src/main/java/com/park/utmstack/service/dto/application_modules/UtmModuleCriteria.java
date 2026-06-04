package com.park.utmstack.service.dto.application_modules;

import com.park.utmstack.domain.application_modules.enums.ModuleName;
import lombok.Getter;
import lombok.Setter;
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
@Setter
@Getter
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

}
