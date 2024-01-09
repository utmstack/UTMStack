package com.park.utmstack.domain;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.application_modules.enums.ModuleName;

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
    @Column(name = "module_name_short")
    private ModuleName moduleNameShort;

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

    public ModuleName getModuleNameShort() {
        return moduleNameShort;
    }

    public void setModuleNameShort(ModuleName moduleNameShort) {
        this.moduleNameShort = moduleNameShort;
    }

    public Boolean getSectionActive() {
        return sectionActive;
    }

    public void setSectionActive(Boolean sectionActive) {
        this.sectionActive = sectionActive;
    }
}
