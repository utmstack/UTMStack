package com.park.utmstack.service.dto.log_analyzer;

import tech.jhipster.service.filter.Filter;
import tech.jhipster.service.filter.InstantFilter;
import tech.jhipster.service.filter.LongFilter;
import tech.jhipster.service.filter.StringFilter;

import java.io.Serializable;

/**
 * Criteria class for the LogAnalyzerQuery entity. This class is used in LogAnalyzerQueryResource to receive all the possible
 * filtering options from the Http GET request parameters. For example the following could be a valid requests:
 * <code> /log-analyzer-queries?id.greaterThan=5&amp;attr1.contains=something&amp;attr2.specified=false</code>
 * As Spring is unable to properly convert the types, unless specific {@link Filter} class are used, we need to use fix type
 * specific filters.
 */
public class LogAnalyzerQueryCriteria implements Serializable {

    private static final long serialVersionUID = 1L;

    private LongFilter id;

    private StringFilter name;

    private StringFilter description;

    private StringFilter owner;

    private InstantFilter creationDate;

    private InstantFilter modificationDate;

    private LongFilter idPattern;

    public LongFilter getId() {
        return id;
    }

    public void setId(LongFilter id) {
        this.id = id;
    }

    public StringFilter getName() {
        return name;
    }

    public void setName(StringFilter name) {
        this.name = name;
    }

    public StringFilter getDescription() {
        return description;
    }

    public void setDescription(StringFilter description) {
        this.description = description;
    }

    public StringFilter getOwner() {
        return owner;
    }

    public void setOwner(StringFilter owner) {
        this.owner = owner;
    }

    public InstantFilter getCreationDate() {
        return creationDate;
    }

    public void setCreationDate(InstantFilter creationDate) {
        this.creationDate = creationDate;
    }

    public InstantFilter getModificationDate() {
        return modificationDate;
    }

    public void setModificationDate(InstantFilter modificationDate) {
        this.modificationDate = modificationDate;
    }

    public LongFilter getIdPattern() {
        return idPattern;
    }

    public void setIdPattern(LongFilter idPattern) {
        this.idPattern = idPattern;
    }
}
