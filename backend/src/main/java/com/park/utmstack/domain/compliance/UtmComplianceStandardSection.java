package com.park.utmstack.domain.compliance;

import org.hibernate.annotations.GenericGenerator;

import jakarta.persistence.*;
import org.springframework.data.domain.Persistable;
import java.io.Serializable;

@Entity
@Table(name = "utm_compliance_standard_section")
public class UtmComplianceStandardSection implements Serializable, Persistable<Long>{
    private static final long serialVersionUID = 1L;

    // Hibernate 6.6 (Boot 3.4.5) changed merge() so that save() on an entity whose id is
    // pre-assigned by the app (this app assigns ids from a sequence / fixed value before
    // persisting) throws ObjectOptimisticLockingFailureException. Tracking whether the instance
    // was loaded from the DB (via @PostLoad) makes Spring Data call persist() for new instances
    // and merge() for existing ones - the pre-6.6 behavior.
    private transient boolean isNew = true;

    @jakarta.persistence.PostLoad
    void utmstackOnLoad() {
        this.isNew = false;
    }

    @Override
    public boolean isNew() {
        return this.isNew;
    }


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
