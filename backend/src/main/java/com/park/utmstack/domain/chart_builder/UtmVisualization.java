package com.park.utmstack.domain.chart_builder;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.chart_builder.types.ChartType;
import com.park.utmstack.domain.chart_builder.types.aggregation.AggregationType;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.hibernate.annotations.GenericGenerator;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.List;

/**
 * A UtmVisualization.
 */
@Entity
@Table(name = "utm_visualization")
public class UtmVisualization implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "name", length = 100, nullable = false, unique = true)
    private String name;

    @Size(max = 255)
    @Column(name = "description")
    private String description;

    @Column(name = "event_type", length = 50, nullable = false)
    private String eventType;

    @Column(name = "created_date")
    private Instant createdDate;

    @Column(name = "modified_date")
    private Instant modifiedDate;

    @Column(name = "user_created", length = 50)
    private String userCreated;

    @Column(name = "user_modified", length = 50)
    private String userModified;

    @Column(name = "chart_config")
    private String chartConfig;

    @Column(name = "chart_action")
    private String chartAction;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    @NotNull
    @Column(name = "id_pattern")
    private Long idPattern;

    @JsonIgnore
    @Column(name = "chart_type")
    private String chType;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private ChartType chartType;

    @JsonIgnore
    @Column(name = "query")
    private String query;

    @JsonIgnore
    @Column(name = "filters")
    private String _filters;

    @JsonIgnore
    @Column(name = "aggregation")
    private String aggregation;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<FilterType> filterType;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private AggregationType aggregationType;

    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "id_pattern", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmIndexPattern pattern;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getEventType() {
        return eventType;
    }

    public void setEventType(String eventType) {
        this.eventType = eventType;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public Instant getModifiedDate() {
        return modifiedDate;
    }

    public void setModifiedDate(Instant modifiedDate) {
        this.modifiedDate = modifiedDate;
    }

    public String getUserCreated() {
        return userCreated;
    }

    public void setUserCreated(String userCreated) {
        this.userCreated = userCreated;
    }

    public String getUserModified() {
        return userModified;
    }

    public void setUserModified(String userModified) {
        this.userModified = userModified;
    }

    public ChartType getChartType() {
        if (StringUtils.hasText(chType))
            chartType = ChartType.valueOf(chType.toUpperCase());
        return chartType;
    }

    public void setChartType(ChartType chartType) {
        if (chartType != null)
            chType = chartType.name();
        this.chartType = chartType;
    }

    public String getQuery() {
        return query;
    }

    public void setQuery(String query) {
        this.query = query;
    }

    public List<FilterType> getFilterType() throws UtmSerializationException {
        filterType = UtilSerializer.jsonDeserializeList(FilterType.class, _filters);
        return filterType;
    }

    public void setFilterType(List<FilterType> filters) throws UtmSerializationException {
        _filters = UtilSerializer.jsonSerialize(filters);
        this.filterType = filters;
    }

    public String getChartConfig() {
        return chartConfig;
    }

    public void setChartConfig(String chartConfig) {
        this.chartConfig = chartConfig;
    }

    public AggregationType getAggregationType() throws UtmSerializationException {
        if (StringUtils.hasText(aggregation))
            aggregationType = UtilSerializer.jsonDeserialize(AggregationType.class, aggregation);
        return aggregationType;
    }

    public void setAggregationType(AggregationType aggregationType) throws UtmSerializationException {
        if (aggregationType != null)
            aggregation = UtilSerializer.jsonSerialize(aggregationType);
        this.aggregationType = aggregationType;
    }

    public String getChartAction() {
        return chartAction;
    }

    public void setChartAction(String chartAction) {
        this.chartAction = chartAction;
    }

    public Long getIdPattern() {
        return idPattern;
    }

    public void setIdPattern(Long idPattern) {
        this.idPattern = idPattern;
    }

    public UtmIndexPattern getPattern() {
        return pattern;
    }

    public void setPattern(UtmIndexPattern pattern) {
        this.pattern = pattern;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
