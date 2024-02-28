package com.park.utmstack.domain.shared_types;

import com.fasterxml.jackson.annotation.JsonIgnoreProperties;
import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;
import com.park.utmstack.util.enums.AlertStatus;
import org.springframework.util.StringUtils;

import java.time.Instant;
import java.time.ZoneId;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Locale;

@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonIgnoreProperties(ignoreUnknown = true)
public class AlertType {
    @JsonProperty("@timestamp")
    private String timestamp;
    private String id;
    private Integer status;
    private AlertStatus statusLabel;
    private String statusObservation;
    private Boolean isIncident;
    private IncidentDetail incidentDetail;
    private String name;
    private String category;
    private Integer severity;
    private String severityLabel;
    private String protocol;
    private String description;
    private String solution;
    private String tactic;
    private List<String> reference;
    private String dataType;
    private String dataSource;
    private Host source;
    private Host destination;
    private List<String> logs;
    private List<String> tags;
    private String notes;
    @JsonProperty("TagRulesApplied")
    private List<Long> tagRulesApplied;

    public String getTimestamp() {
        return timestamp;
    }

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

    public void setTimestamp(String timestamp) {
        this.timestamp = timestamp;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public Integer getStatus() {
        return status;
    }

    public void setStatus(Integer status) {
        this.status = status;
    }

    public AlertStatus getStatusLabel() {
        return statusLabel;
    }

    public void setStatusLabel(AlertStatus statusLabel) {
        this.statusLabel = statusLabel;
    }

    public String getStatusObservation() {
        return statusObservation;
    }

    public void setStatusObservation(String statusObservation) {
        this.statusObservation = statusObservation;
    }

    public Boolean getIncident() {
        return isIncident != null && isIncident;
    }

    public void setIncident(Boolean incident) {
        isIncident = incident;
    }

    public IncidentDetail getIncidentDetail() {
        return incidentDetail;
    }

    public void setIncidentDetail(IncidentDetail incidentDetail) {
        this.incidentDetail = incidentDetail;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getCategory() {
        return category;
    }

    public void setCategory(String category) {
        this.category = category;
    }

    public Integer getSeverity() {
        return severity;
    }

    public void setSeverity(Integer severity) {
        this.severity = severity;
    }

    public String getSeverityLabel() {
        return severityLabel;
    }

    public void setSeverityLabel(String severityLabel) {
        this.severityLabel = severityLabel;
    }

    public String getProtocol() {
        return protocol;
    }

    public void setProtocol(String protocol) {
        this.protocol = protocol;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getSolution() {
        return solution;
    }

    public void setSolution(String solution) {
        this.solution = solution;
    }

    public String getTactic() {
        return tactic;
    }

    public void setTactic(String tactic) {
        this.tactic = tactic;
    }

    public List<String> getReference() {
        return reference;
    }

    public void setReference(List<String> reference) {
        this.reference = reference;
    }

    public String getDataType() {
        return dataType;
    }

    public void setDataType(String dataType) {
        this.dataType = dataType;
    }

    public String getDataSource() {
        return dataSource;
    }

    public void setDataSource(String dataSource) {
        this.dataSource = dataSource;
    }

    public Host getSource() {
        return source;
    }

    public void setSource(Host source) {
        this.source = source;
    }

    public Host getDestination() {
        return destination;
    }

    public void setDestination(Host destination) {
        this.destination = destination;
    }

    public List<String> getLogs() {
        return logs;
    }

    public void setLogs(List<String> logs) {
        this.logs = logs;
    }

    public List<String> getTags() {
        return tags;
    }

    public void setTags(List<String> tags) {
        this.tags = tags;
    }

    public String getNotes() {
        return notes;
    }

    public void setNotes(String notes) {
        this.notes = notes;
    }

    public List<Long> getTagRulesApplied() {
        return tagRulesApplied;
    }

    public void setTagRulesApplied(List<Long> tagRulesApplied) {
        this.tagRulesApplied = tagRulesApplied;
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public static class Host {
        private String user;
        private String host;
        private String ip;
        private Integer port;
        private String country;
        private String countryCode;
        private String city;
        private Float[] coordinates;
        private Integer accuracyRadius;
        private Integer asn;
        private String aso;
        private Boolean isSatelliteProvider;
        private Boolean isAnonymousProxy;

        public String getUser() {
            return user;
        }

        public void setUser(String user) {
            this.user = user;
        }

        public String getHost() {
            return host;
        }

        public void setHost(String host) {
            this.host = host;
        }

        public String getIp() {
            return ip;
        }

        public void setIp(String ip) {
            this.ip = ip;
        }

        public Integer getPort() {
            return port;
        }

        public void setPort(Integer port) {
            this.port = port;
        }

        public String getCountry() {
            return country;
        }

        public void setCountry(String country) {
            this.country = country;
        }

        public String getCountryCode() {
            return countryCode;
        }

        public void setCountryCode(String countryCode) {
            this.countryCode = countryCode;
        }

        public String getCity() {
            return city;
        }

        public void setCity(String city) {
            this.city = city;
        }

        public Float[] getCoordinates() {
            return coordinates;
        }

        public void setCoordinates(Float[] coordinates) {
            this.coordinates = coordinates;
        }

        public Integer getAccuracyRadius() {
            return accuracyRadius;
        }

        public void setAccuracyRadius(Integer accuracyRadius) {
            this.accuracyRadius = accuracyRadius;
        }

        public Integer getAsn() {
            return asn;
        }

        public void setAsn(Integer asn) {
            this.asn = asn;
        }

        public String getAso() {
            return aso;
        }

        public void setAso(String aso) {
            this.aso = aso;
        }

        public Boolean getSatelliteProvider() {
            return isSatelliteProvider;
        }

        public void setSatelliteProvider(Boolean satelliteProvider) {
            isSatelliteProvider = satelliteProvider;
        }

        public Boolean getAnonymousProxy() {
            return isAnonymousProxy;
        }

        public void setAnonymousProxy(Boolean anonymousProxy) {
            isAnonymousProxy = anonymousProxy;
        }
    }

    @JsonIgnoreProperties(ignoreUnknown = true)
    public static class IncidentDetail {
        private String createdBy;
        private String incidentName;

        private Integer incidentId;
        private String creationDate;
        private String source;

        public String getCreatedBy() {
            return createdBy;
        }

        public void setCreatedBy(String createdBy) {
            this.createdBy = createdBy;
        }

        public String getIncidentName() {
            return incidentName;
        }

        public void setIncidentName(String incidentName) {
            this.incidentName = incidentName;
        }

        public Integer getIncidentId() {
            return incidentId;
        }

        public void setIncidentId(Integer incidentId) {
            this.incidentId = incidentId;
        }

        public String getCreationDate() {
            return creationDate;
        }

        public void setCreationDate(String creationDate) {
            this.creationDate = creationDate;
        }

        public String getSource() {
            return source;
        }

        public void setSource(String source) {
            this.source = source;
        }
    }
}
