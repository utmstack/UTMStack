package com.park.utmstack.domain.chart_builder;


import com.park.utmstack.domain.shared_types.enums.DashboardType;
import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;
import java.time.Instant;

/**
 * A UtmDashboard.
 */
@Entity
@Table(name = "utm_dashboard")
public class UtmDashboard implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Size(max = 100)
    @Column(name = "name", length = 100, nullable = false, unique = true)
    private String name;

    @NotNull
    @Size(max = 255)
    @Column(name = "description", length = 255, nullable = false)
    private String description;

    @Column(name = "refresh_time")
    private Long refreshTime;

    @NotNull
    @Column(name = "created_date", nullable = false)
    private Instant createdDate;

    @Column(name = "modified_date", nullable = false)
    private Instant modifiedDate;

    @NotNull
    @Size(max = 50)
    @Column(name = "user_created", nullable = false, length = 50)
    private String userCreated;

    @Size(max = 50)
    @Column(name = "user_modified", length = 50)
    private String userModified;

    @Column(name = "filters")
    private String filters;

    @Column(name = "dashboard_type")
    @Enumerated(EnumType.STRING)
    private DashboardType dashboardType;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getName() {
        return name;
    }

    public UtmDashboard name(String name) {
        this.name = name;
        return this;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public UtmDashboard description(String description) {
        this.description = description;
        return this;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public Instant getCreatedDate() {
        return createdDate;
    }

    public UtmDashboard createdDate(Instant createdDate) {
        this.createdDate = createdDate;
        return this;
    }

    public void setCreatedDate(Instant createdDate) {
        this.createdDate = createdDate;
    }

    public Instant getModifiedDate() {
        return modifiedDate;
    }

    public UtmDashboard modifiedDate(Instant modifiedDate) {
        this.modifiedDate = modifiedDate;
        return this;
    }

    public void setModifiedDate(Instant modifiedDate) {
        this.modifiedDate = modifiedDate;
    }

    public String getUserCreated() {
        return userCreated;
    }

    public UtmDashboard userCreated(String userCreated) {
        this.userCreated = userCreated;
        return this;
    }

    public void setUserCreated(String userCreated) {
        this.userCreated = userCreated;
    }

    public String getUserModified() {
        return userModified;
    }

    public UtmDashboard userModified(String userModified) {
        this.userModified = userModified;
        return this;
    }

    public void setUserModified(String userModified) {
        this.userModified = userModified;
    }

    public String getFilters() {
        return filters;
    }

    public UtmDashboard filters(String filters) {
        this.filters = filters;
        return this;
    }

    public Long getRefreshTime() {
        return refreshTime;
    }

    public void setRefreshTime(Long refreshTime) {
        this.refreshTime = refreshTime;
    }

    public void setFilters(String filters) {
        this.filters = filters;
    }

    public DashboardType getDashboardType() {
        return dashboardType;
    }

    public void setDashboardType(DashboardType dashboardType) {
        this.dashboardType = dashboardType;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
