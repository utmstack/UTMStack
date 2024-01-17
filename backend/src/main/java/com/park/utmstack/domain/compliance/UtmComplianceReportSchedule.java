package com.park.utmstack.domain.compliance;

import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.compliance.types.IndexPatternWithFilterType;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;

import java.io.Serializable;
import java.time.Instant;
import java.util.List;
import javax.persistence.*;
import javax.validation.constraints.Size;

/**
 * A UtmComplianceReportSchedule.
 */
@Entity
@Table(name = "utm_compliance_report_schedule")
public class UtmComplianceReportSchedule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    @Column(name = "id", updatable = false)
    private Long id;

    @Column(name = "user_id")
    private Long userId;

    @Column(name = "compliance_id", nullable = false)
    private Long complianceId;

    @Size(max = 250)
    @Column(name = "schedule_string", nullable = false, length = 250)
    private String scheduleString;

    @Column(name = "url_with_params", nullable = false)
    private String urlWithParams;

    @JsonIgnore
    @Column(name = "filters")
    private String _filters;

    @JsonIgnore
    @Column(name = "last_execution_date", nullable = false)
    private Instant lastExecutionTime;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<IndexPatternWithFilterType> filterDef;


    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "compliance_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmComplianceReportConfig compliance;

    public Long getId() {
        return this.id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getUserId() {
        return this.userId;
    }

    public void setUserId(Long userId) {
        this.userId = userId;
    }

    public String getScheduleString() {
        return this.scheduleString;
    }

    public void setScheduleString(String scheduleString) {
        this.scheduleString = scheduleString;
    }

    public Instant getLastExecutionTime() {
        return lastExecutionTime;
    }

    public void setLastExecutionTime(Instant lastExecutionTime) {
        this.lastExecutionTime = lastExecutionTime;
    }

    public Long getComplianceId() {
        return complianceId;
    }

    public void setComplianceId(Long complianceId) {
        this.complianceId = complianceId;
    }

    public List<IndexPatternWithFilterType> getFilterDef() throws UtmSerializationException {
        filterDef = UtilSerializer.jsonDeserializeList(IndexPatternWithFilterType.class, _filters);
        return filterDef;
    }

    public void setFilterDef(List<IndexPatternWithFilterType> filters) throws UtmSerializationException {
        _filters = UtilSerializer.jsonSerialize(filters);
        this.filterDef = filters;
    }

    public UtmComplianceReportConfig getCompliance() {
        return compliance;
    }

    public void setCompliance(UtmComplianceReportConfig compliance) {
        this.compliance = compliance;
    }

    public String getUrlWithParams() {
        return urlWithParams;
    }

    public void setUrlWithParams(String urlWithParams) {
        this.urlWithParams = urlWithParams;
    }

    @Override
    public boolean equals(Object o) {
        if (o instanceof UtmComplianceReportSchedule) {
            return this.userId == ((UtmComplianceReportSchedule) o).userId
                && this.scheduleString.compareTo(((UtmComplianceReportSchedule) o).getScheduleString()) == 0
                && this.urlWithParams.compareTo(((UtmComplianceReportSchedule) o).getUrlWithParams()) == 0
                && this.complianceId == ((UtmComplianceReportSchedule) o).getComplianceId();
        }
        return false;
    }

}
