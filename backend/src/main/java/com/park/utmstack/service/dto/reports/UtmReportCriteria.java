package com.park.utmstack.service.dto.reports;

import com.park.utmstack.domain.reports.enums.ReportTypeEnum;
import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the UtmReports entity. This class is used in UtmReportsResource to
 * receive all the possible filtering options from the Http GET request parameters.
 * For example the following could be a valid requests:
 * <code> /utm-reports?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use
 * fix type specific filters.
 */
public class UtmReportCriteria implements Serializable {

    /**
     * Class for filtering ModulesNameShort
     */
    public static class ReportTypeFilter extends Filter<ReportTypeEnum> {
    }

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter repName;

    private StringFilter repDescription;

    private LongFilter reportSectionId;

    private LongFilter dashboardId;

    private StringFilter creationUser;

    private InstantFilter creationDate;

    private StringFilter modificationUser;

    private InstantFilter modificationDate;

    private ReportTypeFilter repType;

    private StringFilter repModule;

    private StringFilter repShortName;

    private StringFilter repHttpMethod;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getRepName() {
        return repName;
    }

    public void setRepName(StringFilter repName) {
        this.repName = repName;
    }

    public StringFilter getRepDescription() {
        return repDescription;
    }

    public void setRepDescription(StringFilter repDescription) {
        this.repDescription = repDescription;
    }

    public LongFilter getReportSectionId() {
        return reportSectionId;
    }

    public void setReportSectionId(LongFilter reportSectionId) {
        this.reportSectionId = reportSectionId;
    }

    public LongFilter getDashboardId() {
        return dashboardId;
    }

    public void setDashboardId(LongFilter dashboardId) {
        this.dashboardId = dashboardId;
    }

    public StringFilter getCreationUser() {
        return creationUser;
    }

    public void setCreationUser(StringFilter creationUser) {
        this.creationUser = creationUser;
    }

    public InstantFilter getCreationDate() {
        return creationDate;
    }

    public void setCreationDate(InstantFilter creationDate) {
        this.creationDate = creationDate;
    }

    public StringFilter getModificationUser() {
        return modificationUser;
    }

    public void setModificationUser(StringFilter modificationUser) {
        this.modificationUser = modificationUser;
    }

    public InstantFilter getModificationDate() {
        return modificationDate;
    }

    public void setModificationDate(InstantFilter modificationDate) {
        this.modificationDate = modificationDate;
    }

    public ReportTypeFilter getRepType() {
        return repType;
    }

    public void setRepType(ReportTypeFilter repType) {
        this.repType = repType;
    }

    public StringFilter getRepModule() {
        return repModule;
    }

    public void setRepModule(StringFilter repModule) {
        this.repModule = repModule;
    }

    public StringFilter getRepShortName() {
        return repShortName;
    }

    public void setRepShortName(StringFilter repShortName) {
        this.repShortName = repShortName;
    }

    public StringFilter getRepHttpMethod() {
        return repHttpMethod;
    }

    public void setRepHttpMethod(StringFilter repHttpMethod) {
        this.repHttpMethod = repHttpMethod;
    }
}
