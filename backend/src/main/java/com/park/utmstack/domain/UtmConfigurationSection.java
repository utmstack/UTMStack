package com.park.utmstack.domain;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.shared_types.enums.SectionType;

import javax.persistence.*;
import java.io.Serializable;

/**
 * A UtmConfigurationSection.
 */
@Entity
@Table(name = "utm_configuration_section")
public class UtmConfigurationSection implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "section")
    private String section;

    @Column(name = "description")
    private String description;

    @Enumerated(EnumType.STRING)
    @Column(name = "short_name")
    private SectionType shortName;

    @JsonIgnore
    @Column(name = "section_active")
    private Boolean sectionActive;


    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getSection() {
        return section;
    }

    public void setSection(String section) {
        this.section = section;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public Boolean getSectionActive() {
        return sectionActive;
    }

    public void setSectionActive(Boolean sectionActive) {
        this.sectionActive = sectionActive;
    }

    public SectionType getShortName() {
        return shortName;
    }

    public void setShortName(SectionType sectionType) {
        this.shortName = sectionType;
    }
}
