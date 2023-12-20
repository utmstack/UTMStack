package com.park.utmstack.domain.compliance;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.chart_builder.UtmDashboard;
import com.park.utmstack.domain.chart_builder.UtmDashboardVisualization;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.compliance.enums.ComplianceType;
import com.park.utmstack.domain.compliance.types.RequestParamFilter;
import com.park.utmstack.domain.shared_types.DataColumn;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import java.io.Serializable;
import java.util.List;

/**
 * A ComplianceTemplate.
 */
@Entity
@Table(name = "utm_compliance_report_config")
public class UtmComplianceReportConfig implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "config_solution")
    private String configSolution;

    @JsonIgnore
    @Column(name = "config_report_columns")
    private String configReportColumns;

    @JsonIgnore
    @Column(name = "config_report_req_body")
    private String configReportReqBody;

    @JsonIgnore
    @Column(name = "config_report_req_params")
    private String configReportReqParams;

    @Column(name = "config_report_resource_url")
    private String configReportResourceUrl;

    @Column(name = "config_report_request_type")
    private String configReportRequestType;

    @Column(name = "config_report_pageable")
    private Boolean configReportPageable;

    @Column(name = "config_report_filter_by_time")
    private Boolean configReportFilterByTime;

    @Column(name = "config_report_data_origin", length = 50)
    private String configReportDataOrigin;

    @Column(name = "config_report_export_csv_url", length = 50)
    private String configReportExportCsvUrl;

    @Column(name = "standard_section_id")
    private Long standardSectionId;

    @Column(name = "config_report_editable")
    private Boolean configReportEditable;

    @Column(name = "dashboard_id")
    private Long dashboardId;

    @Enumerated(EnumType.STRING)
    @Column(name = "config_type")
    private ComplianceType configType;

    @Column(name = "config_url")
    private String configUrl;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<UtmDashboardVisualization> dashboard;

    @ManyToOne
    @JoinColumn(name = "standard_section_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmComplianceStandardSection section;

    @OneToOne
    @JoinColumn(name = "dashboard_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmDashboard associatedDashboard;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<DataColumn> columns;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<FilterType> requestBodyFilters;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<RequestParamFilter> requestParamFilters;


    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getConfigSolution() {
        return configSolution;
    }

    public void setConfigSolution(String configSolution) {
        this.configSolution = configSolution;
    }

    public String getConfigReportColumns() {
        return configReportColumns;
    }

    public void setConfigReportColumns(String configReportColumns) {
        this.configReportColumns = configReportColumns;
    }

    public String getConfigReportReqBody() {
        return configReportReqBody;
    }

    public void setConfigReportReqBody(String configReportReqBody) {
        this.configReportReqBody = configReportReqBody;
    }

    public String getConfigReportReqParams() {
        return configReportReqParams;
    }

    public void setConfigReportReqParams(String configReportReqParams) {
        this.configReportReqParams = configReportReqParams;
    }

    public String getConfigReportResourceUrl() {
        return configReportResourceUrl;
    }

    public void setConfigReportResourceUrl(String configReportResourceUrl) {
        this.configReportResourceUrl = configReportResourceUrl;
    }

    public String getConfigReportRequestType() {
        return configReportRequestType;
    }

    public void setConfigReportRequestType(String configReportRequestType) {
        this.configReportRequestType = configReportRequestType;
    }

    public Boolean getConfigReportPageable() {
        return configReportPageable;
    }

    public void setConfigReportPageable(Boolean configReportPageable) {
        this.configReportPageable = configReportPageable;
    }

    public Boolean getConfigReportFilterByTime() {
        return configReportFilterByTime;
    }

    public void setConfigReportFilterByTime(Boolean configReportFilterByTime) {
        this.configReportFilterByTime = configReportFilterByTime;
    }

    public String getConfigReportDataOrigin() {
        return configReportDataOrigin;
    }

    public void setConfigReportDataOrigin(String configReportDataOrigin) {
        this.configReportDataOrigin = configReportDataOrigin;
    }

    public String getConfigReportExportCsvUrl() {
        return configReportExportCsvUrl;
    }

    public void setConfigReportExportCsvUrl(String configReportExportCsvUrl) {
        this.configReportExportCsvUrl = configReportExportCsvUrl;
    }

    public Long getStandardSectionId() {
        return standardSectionId;
    }

    public void setStandardSectionId(Long standardSectionId) {
        this.standardSectionId = standardSectionId;
    }

    public Boolean getConfigReportEditable() {
        return configReportEditable;
    }

    public void setConfigReportEditable(Boolean configReportEditable) {
        this.configReportEditable = configReportEditable;
    }

    public UtmComplianceStandardSection getSection() {
        return section;
    }

    public void setSection(UtmComplianceStandardSection section) {
        this.section = section;
    }

    public List<DataColumn> getColumns() throws UtmSerializationException {
        return StringUtils.hasText(configReportColumns) ? UtilSerializer.jsonDeserializeList(DataColumn.class, configReportColumns) : columns;
    }

    public void setColumns(List<DataColumn> columns) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(columns))
            this.configReportColumns = null;
        else
            this.configReportColumns = UtilSerializer.jsonSerialize(columns);
        this.columns = columns;
    }

    public List<FilterType> getRequestBodyFilters() throws UtmSerializationException {
        if (StringUtils.hasText(configReportReqBody))
            requestBodyFilters = UtilSerializer.jsonDeserializeList(FilterType.class, configReportReqBody);
        return requestBodyFilters;
    }

    public void setRequestBodyFilters(List<FilterType> requestBodyFilters) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(requestBodyFilters))
            this.configReportReqBody = null;
        else
            this.configReportReqBody = UtilSerializer.jsonSerialize(requestBodyFilters);
        this.requestBodyFilters = requestBodyFilters;
    }

    public List<RequestParamFilter> getRequestParamFilters() throws UtmSerializationException {
        if (StringUtils.hasText(configReportReqParams))
            requestParamFilters = UtilSerializer.jsonDeserializeList(RequestParamFilter.class, configReportReqParams);
        return requestParamFilters;
    }

    public void setRequestParamFilters(List<RequestParamFilter> requestParamFilters) throws UtmSerializationException {
        if (CollectionUtils.isEmpty(requestParamFilters))
            this.configReportReqParams = null;
        else
            this.configReportReqParams = UtilSerializer.jsonSerialize(requestParamFilters);
        this.requestParamFilters = requestParamFilters;
    }

    public Long getDashboardId() {
        return dashboardId;
    }

    public void setDashboardId(Long dashboardId) {
        this.dashboardId = dashboardId;
    }

    public List<UtmDashboardVisualization> getDashboard() {
        return dashboard;
    }

    public void setDashboard(List<UtmDashboardVisualization> dashboard) {
        this.dashboard = dashboard;
    }

    public ComplianceType getConfigType() {
        return configType;
    }

    public void setConfigType(ComplianceType configType) {
        this.configType = configType;
    }

    public String getConfigUrl() {
        return configUrl;
    }

    public void setConfigUrl(String configUrl) {
        this.configUrl = configUrl;
    }

    public UtmDashboard getAssociatedDashboard() {
        return associatedDashboard;
    }

    public void setAssociatedDashboard(UtmDashboard associatedDashboard) {
        this.associatedDashboard = associatedDashboard;
    }
}
