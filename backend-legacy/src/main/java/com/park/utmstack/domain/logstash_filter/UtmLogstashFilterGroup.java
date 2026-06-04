package com.park.utmstack.domain.logstash_filter;


import org.hibernate.annotations.GenericGenerator;

import javax.persistence.*;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Size;
import java.io.Serializable;

/**
 * A UtmLogstashFilterGroup.
 */
@Entity
@Table(name = "utm_logstash_filter_group")
public class UtmLogstashFilterGroup implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GenericGenerator(name = "CustomIdentityGenerator", strategy = "com.park.utmstack.util.CustomIdentityGenerator")
    @GeneratedValue(generator = "CustomIdentityGenerator")
    private Long id;

    @NotNull
    @Size(max = 150)
    @Column(name = "group_name", length = 150, nullable = false, unique = true)
    private String groupName;

    @Column(name = "group_description")
    private String groupDescription;

    @Column(name = "system_owner")
    private Boolean systemOwner;

    public UtmLogstashFilterGroup() {
    }

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getGroupName() {
        return groupName;
    }

    public UtmLogstashFilterGroup groupName(String groupName) {
        this.groupName = groupName;
        return this;
    }

    public void setGroupName(String groupName) {
        this.groupName = groupName;
    }

    public String getGroupDescription() {
        return groupDescription;
    }

    public UtmLogstashFilterGroup groupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
        return this;
    }

    public void setGroupDescription(String groupDescription) {
        this.groupDescription = groupDescription;
    }

    public Boolean getSystemOwner() {
        return systemOwner;
    }

    public void setSystemOwner(Boolean systemOwner) {
        this.systemOwner = systemOwner;
    }
}
