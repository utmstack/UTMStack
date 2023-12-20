package com.park.utmstack.domain.logstash_filter;


import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotBlank;
import java.io.Serializable;

/**
 * A UtmLogstashFilter.
 */
@Entity
@Table(name = "utm_logstash_filter")
public class UtmLogstashFilter implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @Column(name = "filter_name")
    private String filterName;

    @NotBlank
    @Column(name = "logstash_filter")
    private String logstashFilter;

    @Column(name = "filter_group_id")
    private Long filterGroupId;

    @ManyToOne(fetch = FetchType.EAGER)
    @JoinColumn(name = "filter_group_id", referencedColumnName = "id", insertable = false, updatable = false)
    private UtmLogstashFilterGroup group;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    @Column(name = "is_active")
    private Boolean isActive = true;

    @Column(name = "module_name")
    private String moduleName;

    @Column(name = "filter_version")
    private String filterVersion;

    public UtmLogstashFilter() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getFilterName() {
        return filterName;
    }

    public void setFilterName(String filterName) {
        this.filterName = filterName;
    }

    public String getLogstashFilter() {
        return logstashFilter;
    }

    public UtmLogstashFilter logstashFilter(String logstashFilter) {
        this.logstashFilter = logstashFilter;
        return this;
    }

    public void setLogstashFilter(String logstashFilter) {
        this.logstashFilter = logstashFilter;
    }

    public Long getFilterGroupId() {
        return filterGroupId;
    }

    public void setFilterGroupId(Long filterGroupId) {
        this.filterGroupId = filterGroupId;
    }

    public UtmLogstashFilterGroup getGroup() {
        return group;
    }

    public void setGroup(UtmLogstashFilterGroup group) {
        this.group = group;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }

    public Boolean getActive() {
        return isActive;
    }

    public void setActive(Boolean active) {
        isActive = active;
    }

    public String getModuleName() {
        return moduleName;
    }

    public void setModuleName(String moduleName) {
        this.moduleName = moduleName;
    }

    public String getFilterVersion() {
        return filterVersion;
    }

    public void setFilterVersion(String filterVersion) {
        this.filterVersion = filterVersion;
    }
}
