package com.park.utmstack.domain.correlation.config;

import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Clock;
import java.time.Instant;

/**
 * UtmRegexPattern entity template, used for filters processed by correlation rule engine.
 */
@Entity
@Table(name = "utm_regex_pattern")
public class UtmRegexPattern implements Serializable {
    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    @Column(name = "id", updatable = false)
    private Long id;

    @Size(max = 100)
    @Column(name = "pattern_id", length = 100, nullable = false)
    private String patternId;

    @Column(name = "pattern_description")
    private String patternDescription;

    @Column(name = "pattern_definition", nullable = false)
    private String patternDefinition;

    @Column(name = "system_owner", nullable = false)
    private Boolean systemOwner;

    @Column(name = "last_update")
    private Instant lastUpdate;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getPatternId() {
        return patternId;
    }

    public void setPatternId(String patternId) {
        this.patternId = patternId;
    }

    public String getPatternDescription() {
        return patternDescription;
    }

    public void setPatternDescription(String patternDescription) {
        this.patternDescription = patternDescription;
    }

    public String getPatternDefinition() {
        return patternDefinition;
    }

    public void setPatternDefinition(String patternDefinition) {
        this.patternDefinition = patternDefinition;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public Instant getLastUpdate() {
        return lastUpdate;
    }

    public void setLastUpdate() {
        this.lastUpdate = Instant.now(Clock.systemUTC());
    }
}
