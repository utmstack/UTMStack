package com.park.utmstack.domain.reports;


import org.hibernate.annotations.GenericGenerator;

import jakarta.persistence.*;
import org.springframework.data.domain.Persistable;
import jakarta.validation.constraints.NotNull;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmReportSection.
 */
@Entity
@Table(name = "utm_report_section")
public class UtmReportSection implements Serializable, Persistable<Long>{

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

    @Column(name = "rep_sec_name")
    private String repSecName;

    @NotNull
    @Column(name = "rep_sec_description", nullable = false)
    private String repSecDescription;

    @Column(name = "rep_sec_system")
    private Boolean repSecSystem;

    @Column(name = "creation_user")
    private String creationUser;

    @Column(name = "creation_date")
    private Instant creationDate;

    @Column(name = "modification_user")
    private String modificationUser;

    @Column(name = "modification_date")
    private Instant modificationDate;

    public UtmReportSection() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getRepSecName() {
        return repSecName;
    }

    public UtmReportSection repSecName(String repSecName) {
        this.repSecName = repSecName;
        return this;
    }

    public void setRepSecName(String repSecName) {
        this.repSecName = repSecName;
    }

    public String getRepSecDescription() {
        return repSecDescription;
    }

    public UtmReportSection repSecDescription(String repSecDescription) {
        this.repSecDescription = repSecDescription;
        return this;
    }

    public void setRepSecDescription(String repSecDescription) {
        this.repSecDescription = repSecDescription;
    }

    public String getCreationUser() {
        return creationUser;
    }

    public UtmReportSection creationUser(String creationUser) {
        this.creationUser = creationUser;
        return this;
    }

    public void setCreationUser(String creationUser) {
        this.creationUser = creationUser;
    }

    public Instant getCreationDate() {
        return creationDate;
    }

    public UtmReportSection creationDate(Instant creationDate) {
        this.creationDate = creationDate;
        return this;
    }

    public void setCreationDate(Instant creationDate) {
        this.creationDate = creationDate;
    }

    public String getModificationUser() {
        return modificationUser;
    }

    public UtmReportSection modificationUser(String modificationUser) {
        this.modificationUser = modificationUser;
        return this;
    }

    public void setModificationUser(String modificationUser) {
        this.modificationUser = modificationUser;
    }

    public Instant getModificationDate() {
        return modificationDate;
    }

    public UtmReportSection modificationDate(Instant modificationDate) {
        this.modificationDate = modificationDate;
        return this;
    }

    public void setModificationDate(Instant modificationDate) {
        this.modificationDate = modificationDate;
    }

    public Boolean getRepSecSystem() {
        return repSecSystem;
    }

    public void setRepSecSystem(Boolean repSecSystem) {
        this.repSecSystem = repSecSystem;
    }
}
