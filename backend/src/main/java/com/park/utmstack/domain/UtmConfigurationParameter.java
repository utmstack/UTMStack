package com.park.utmstack.domain;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmConfigurationParameter.
 */
@Entity
@Table(name = "utm_configuration_parameter")
public class UtmConfigurationParameter implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @NotNull
    @Column(name = "section_id", nullable = false)
    private Long sectionId;

    @NotNull
    @Size(max = 100)
    @Column(name = "conf_param_short", length = 100, nullable = false)
    private String confParamShort;

    @Column(name = "conf_param_large")
    private String confParamLarge;

    @Column(name = "conf_param_description")
    private String confParamDescription;

    @Column(name = "conf_param_value")
    private String confParamValue;

    @Column(name = "conf_param_regexp")
    private String confParamRegexp;

    @Column(name = "conf_param_required")
    private Boolean confParamRequired;

    @Column(name = "conf_param_datatype")
    private String confParamDatatype;

    @Column(name = "modification_time")
    private Instant modificationTime;

    @Column(name = "modification_user")
    private String modificationUser;

    @Column(name = "conf_param_option")
    private String confParamOption;
    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "section_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmConfigurationSection section;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getSectionId() {
        return sectionId;
    }

    public void setSectionId(Long sectionId) {
        this.sectionId = sectionId;
    }

    public String getConfParamShort() {
        return confParamShort;
    }

    public void setConfParamShort(String confParamShort) {
        this.confParamShort = confParamShort;
    }

    public String getConfParamLarge() {
        return confParamLarge;
    }

    public void setConfParamLarge(String confParamLarge) {
        this.confParamLarge = confParamLarge;
    }

    public String getConfParamDescription() {
        return confParamDescription;
    }

    public void setConfParamDescription(String confParamDescription) {
        this.confParamDescription = confParamDescription;
    }

    public String getConfParamValue() {
        return confParamValue;
    }

    public void setConfParamValue(String confParamValue) {
        this.confParamValue = confParamValue;
    }

    public Boolean getConfParamRequired() {
        return confParamRequired;
    }

    public void setConfParamRequired(Boolean confParamRequired) {
        this.confParamRequired = confParamRequired;
    }

    public String getConfParamDatatype() {
        return confParamDatatype;
    }

    public void setConfParamDatatype(String confParamDatatype) {
        this.confParamDatatype = confParamDatatype;
    }

    public Instant getModificationTime() {
        return modificationTime;
    }

    public void setModificationTime(Instant modificationTime) {
        this.modificationTime = modificationTime;
    }

    public String getModificationUser() {
        return modificationUser;
    }

    public void setModificationUser(String modificationUser) {
        this.modificationUser = modificationUser;
    }

    public UtmConfigurationSection getSection() {
        return section;
    }

    public void setSection(UtmConfigurationSection section) {
        this.section = section;
    }

    public static long getSerialVersionUID() {
        return serialVersionUID;
    }

    public String getConfParamOption() {
        return confParamOption;
    }

    public void setConfParamOption(String confParamOption) {
        this.confParamOption = confParamOption;
    }

    public String getConfParamRegexp() {
        return confParamRegexp;
    }

    public void setConfParamRegexp(String confParamRegexp) {
        this.confParamRegexp = confParamRegexp;
    }
}
