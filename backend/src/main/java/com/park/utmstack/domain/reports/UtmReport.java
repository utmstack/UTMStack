package com.park.utmstack.domain.reports;


import com.park.utmstack.domain.reports.enums.ReportTypeEnum;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmReports.
 */
@Entity
@Table(name = "utm_report")
public class UtmReport implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Size(max = 255)
    @Column(name = "rep_name")
    private String repName;

    @Size(max = 255)
    @Column(name = "rep_description")
    private String repDescription;

    @Column(name = "report_section_id")
    private Long reportSectionId;

    @Column(name = "dashboard_id")
    private Long dashboardId;

    @Column(name = "creation_user")
    private String creationUser;

    @Column(name = "creation_date")
    private Instant creationDate;

    @Column(name = "modification_user")
    private String modificationUser;

    @Column(name = "modification_date")
    private Instant modificationDate;

    @Column(name = "rep_url")
    private String repUrl;

    @Enumerated(value = EnumType.STRING)
    @Column(name = "rep_type")
    private ReportTypeEnum repType;

    @Column(name = "rep_module")
    private String repModule;

    @Column(name = "rep_short_name")
    private String repShortName;

    @Column(name = "rep_http_method")
    private String repHttpMethod;

    public UtmReport() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getRepName() {
        return repName;
    }

    public UtmReport repName(String repName) {
        this.repName = repName;
        return this;
    }

    public void setRepName(String repName) {
        this.repName = repName;
    }

    public String getRepDescription() {
        return repDescription;
    }

    public UtmReport repDescription(String repDescription) {
        this.repDescription = repDescription;
        return this;
    }

    public void setRepDescription(String repDescription) {
        this.repDescription = repDescription;
    }

    public Long getReportSectionId() {
        return reportSectionId;
    }

    public UtmReport reportSectionId(Long reportSectionId) {
        this.reportSectionId = reportSectionId;
        return this;
    }

    public void setReportSectionId(Long reportSectionId) {
        this.reportSectionId = reportSectionId;
    }

    public Long getDashboardId() {
        return dashboardId;
    }

    public UtmReport dashboardId(Long dashboardId) {
        this.dashboardId = dashboardId;
        return this;
    }

    public void setDashboardId(Long dashboardId) {
        this.dashboardId = dashboardId;
    }

    public String getCreationUser() {
        return creationUser;
    }

    public UtmReport creationUser(String creationUser) {
        this.creationUser = creationUser;
        return this;
    }

    public void setCreationUser(String creationUser) {
        this.creationUser = creationUser;
    }

    public Instant getCreationDate() {
        return creationDate;
    }

    public UtmReport creationDate(Instant creationDate) {
        this.creationDate = creationDate;
        return this;
    }

    public void setCreationDate(Instant creationDate) {
        this.creationDate = creationDate;
    }

    public String getModificationUser() {
        return modificationUser;
    }

    public UtmReport modificationUser(String modificationUser) {
        this.modificationUser = modificationUser;
        return this;
    }

    public void setModificationUser(String modificationUser) {
        this.modificationUser = modificationUser;
    }

    public Instant getModificationDate() {
        return modificationDate;
    }

    public UtmReport modificationDate(Instant modificationDate) {
        this.modificationDate = modificationDate;
        return this;
    }

    public void setModificationDate(Instant modificationDate) {
        this.modificationDate = modificationDate;
    }

    public String getRepUrl() {
        return repUrl;
    }

    public void setRepUrl(String repUrl) {
        this.repUrl = repUrl;
    }

    public ReportTypeEnum getRepType() {
        return repType;
    }

    public void setRepType(ReportTypeEnum repType) {
        this.repType = repType;
    }

    public String getRepModule() {
        return repModule;
    }

    public void setRepModule(String repModule) {
        this.repModule = repModule;
    }

    public String getRepShortName() {
        return repShortName;
    }

    public void setRepShortName(String repShortName) {
        this.repShortName = repShortName;
    }

    public String getRepHttpMethod() {
        return repHttpMethod;
    }

    public void setRepHttpMethod(String repHttpMethod) {
        this.repHttpMethod = repHttpMethod;
    }
}
