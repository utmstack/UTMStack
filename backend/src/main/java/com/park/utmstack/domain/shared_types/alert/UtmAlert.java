package com.park.utmstack.domain.shared_types.alert;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.util.enums.AlertStatus;
import lombok.Getter;
import lombok.Setter;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Locale;

@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
@Getter
@Setter
public class UtmAlert {
    @JsonProperty("@timestamp")
    private String timestamp;

    @JsonProperty("id")
    private String id;

    @JsonProperty("parentId")
    private String parentId;

    @JsonProperty("status")
    private Integer status;

    @JsonProperty("statusLabel")
    private AlertStatus statusLabel;

    @JsonProperty("statusObservation")
    private String statusObservation;

    @JsonProperty("isIncident")
    private Boolean isIncident;

    @JsonProperty("incidentDetail")
    private IncidentDetail incidentDetail;

    @JsonProperty("name")
    private String name;

    @JsonProperty("category")
    private String category;

    @JsonProperty("severity")
    private Integer severity;

    @JsonProperty("severityLabel")
    private String severityLabel;

    @JsonProperty("description")
    private String description;

    @JsonProperty("solution")
    private String solution;

    @JsonProperty("technique")
    private String technique;

    @JsonProperty("reference")
    private List<String> reference;

    @JsonProperty("dataType")
    private String dataType;

    @JsonProperty("impact")
    private Impact impact;

    @JsonProperty("impactScore")
    private Integer impactScore;

    @JsonProperty("dataSource")
    private String dataSource;

    @JsonProperty("adversary")
    private Side adversary;

    @JsonProperty("target")
    private Side target;

    @JsonProperty("events")
    private List<Event> events;

    @JsonProperty("lastEvent")
    private Event lastEvent;

    @JsonProperty("tags")
    private List<String> tags;

    @JsonProperty("notes")
    private String notes;

    @JsonProperty("tagRulesApplied")
    private List<Long> tagRulesApplied;

    @JsonProperty("deduplicatedBy")
    private List<String> deduplicatedBy;

    @JsonProperty("logs")
    private List<String> logs;

    private String assetGroupName;

    private Long assetGroupId;

    public Instant getTimestampAsInstant() {
        if (StringUtils.hasText(timestamp))
            return Instant.parse(timestamp);
        return null;
    }

    public String getTimestampFormatted() {
        try {
            if (!StringUtils.hasText(timestamp))
                return null;
            return DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss").withLocale(Locale.getDefault()).withZone(
                ZoneId.systemDefault()).format(Instant.parse(timestamp));
        } catch (Exception e) {
            return null;
        }
    }

    public Boolean getIncident() {
        return isIncident != null && isIncident;
    }
}
