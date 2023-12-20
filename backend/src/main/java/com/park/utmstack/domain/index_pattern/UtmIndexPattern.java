package com.park.utmstack.domain.index_pattern;


import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmIndexPattern.
 */
@Entity
@Table(name = "utm_index_pattern")
public class UtmIndexPattern implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "pattern", length = 100, nullable = false, unique = true)
    private String pattern;

    @Column(name = "pattern_module", length = 500)
    private String patternModule;

    @Column(name = "pattern_system")
    private Boolean patternSystem;

    @Column(name = "is_active")
    private Boolean isActive = true;

    public UtmIndexPattern() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getPattern() {
        return pattern;
    }

    public void setPattern(String pattern) {
        this.pattern = pattern;
    }

    public String getPatternModule() {
        return patternModule;
    }

    public void setPatternModule(String patternModule) {
        this.patternModule = patternModule;
    }

    public Boolean getPatternSystem() {
        return patternSystem;
    }

    public void setPatternSystem(Boolean patternSystem) {
        this.patternSystem = patternSystem;
    }

    public Boolean getActive() {
        return isActive;
    }

    public void setActive(Boolean active) {
        isActive = active;
    }
}
