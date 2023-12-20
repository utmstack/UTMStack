package com.park.utmstack.domain.compliance;

import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import java.io.Serializable;

@Entity
@Table(name = "utm_compliance_standard_section")
public class UtmComplianceStandardSection implements Serializable {
    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "standard_id")
    private Long standardId;

    @Column(name = "standard_section_name")
    private String standardSectionName;

    @Column(name = "standard_section_description")
    private String standardSectionDescription;

    @ManyToOne
    @JoinColumn(name = "standard_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmComplianceStandard standard;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public Long getStandardId() {
        return standardId;
    }

    public void setStandardId(Long standardId) {
        this.standardId = standardId;
    }

    public String getStandardSectionName() {
        return standardSectionName;
    }

    public void setStandardSectionName(String standardSectionName) {
        this.standardSectionName = standardSectionName;
    }

    public String getStandardSectionDescription() {
        return standardSectionDescription;
    }

    public void setStandardSectionDescription(String standardSectionDescription) {
        this.standardSectionDescription = standardSectionDescription;
    }

    public UtmComplianceStandard getStandard() {
        return standard;
    }

    public void setStandard(UtmComplianceStandard standard) {
        this.standard = standard;
    }
}
