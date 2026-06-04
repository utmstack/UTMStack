package com.park.utmstack.domain.log_analyzer;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.fasterxml.jackson.databind.annotation.JsonDeserialize;
import com.fasterxml.jackson.databind.annotation.JsonSerialize;
import com.park.utmstack.domain.chart_builder.types.query.FilterType;
import com.park.utmstack.domain.index_pattern.UtmIndexPattern;
import com.park.utmstack.domain.shared_types.DataColumn;
import com.park.utmstack.util.UtilSerializer;
import com.park.utmstack.util.exceptions.UtmSerializationException;
import org.springframework.util.CollectionUtils;
import org.springframework.util.StringUtils;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;
import java.util.List;

/**
 * A LogAnalyzerQuery.
 */
@Entity
@Table(name = "utm_log_analyzer_query")
public class LogAnalyzerQuery implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "la_name", length = 100, nullable = false)
    private String name;

    @Size(max = 255)
    @Column(name = "la_description")
    private String description;

    @Column(name = "la_owner")
    private String owner;

    @Column(name = "la_creation_date")
    private Instant creationDate;

    @Column(name = "la_modification_date")
    private Instant modificationDate;

    @JsonIgnore
    @Column(name = "la_columns")
    private String columns;

    @JsonIgnore
    @Column(name = "la_filters")
    private String filters;

    @Column(name = "la_data_origin")
    private String dataOrigin;

    @Column(name = "id_pattern")
    private Long idPattern;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<DataColumn> columnsType;

    @Transient
    @JsonSerialize
    @JsonDeserialize
    private List<FilterType> filtersType;

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

    public LogAnalyzerQuery name(String name) {
        this.name = name;
        return this;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public LogAnalyzerQuery description(String description) {
        this.description = description;
        return this;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getOwner() {
        return owner;
    }

    public void setOwner(String owner) {
        this.owner = owner;
    }

    public Instant getCreationDate() {
        return creationDate;
    }

    public void setCreationDate(Instant creationDate) {
        this.creationDate = creationDate;
    }

    public Instant getModificationDate() {
        return modificationDate;
    }

    public void setModificationDate(Instant modificationDate) {
        this.modificationDate = modificationDate;
    }

    public List<DataColumn> getColumnsType() throws UtmSerializationException {
        if (StringUtils.hasText(columns))
            columnsType = UtilSerializer.jsonDeserializeList(DataColumn.class, columns);
        return columnsType;
    }

    public void setColumnsType(List<DataColumn> columnsType) throws UtmSerializationException {
        if (!CollectionUtils.isEmpty(columnsType))
            columns = UtilSerializer.jsonSerialize(columnsType);
        this.columnsType = columnsType;
    }

    public List<FilterType> getFiltersType() throws UtmSerializationException {
        if (StringUtils.hasText(filters))
            filtersType = UtilSerializer.jsonDeserializeList(FilterType.class, filters);
        return filtersType;
    }

    public void setFiltersType(List<FilterType> filtersType) throws UtmSerializationException {
        if (!CollectionUtils.isEmpty(filtersType))
            filters = UtilSerializer.jsonSerialize(filtersType);
        this.filtersType = filtersType;
    }

    public String getDataOrigin() {
        return dataOrigin;
    }

    public void setDataOrigin(String dataOrigin) {
        this.dataOrigin = dataOrigin;
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
}
