package com.park.utmstack.domain.application_modules;


import com.fasterxml.jackson.annotation.JsonIgnore;
import com.park.utmstack.domain.UtmServer;
import com.park.utmstack.domain.application_modules.enums.ModuleName;
import com.park.utmstack.domain.logstash_filter.UtmLogstashFilter;
import lombok.Data;

import javax.persistence.*;
import java.io.Serializable;
import java.util.HashSet;
import java.util.Set;

/**
 * A UtmModule.
 */
@Entity
@Data
@Table(name = "utm_module")
public class UtmModule implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "server_id")
    private Long serverId;

    @Column(name = "pretty_name")
    private String prettyName;

    @Convert(converter = com.park.utmstack.domain.application_modules.enums.ModuleNameConverter.class)
    @Column(name = "module_name")
    private ModuleName moduleName;

    @Column(name = "module_description")
    private String moduleDescription;

    @Column(name = "module_active")
    private Boolean moduleActive;

    @Column(name = "module_icon")
    private String moduleIcon;

    @Column(name = "module_category")
    private String moduleCategory;

    @Column(name = "lite_version")
    private Boolean liteVersion;

    @Column(name = "needs_restart")
    private Boolean needsRestart;

    @Column(name = "is_activatable")
    private Boolean isActivatable;

    @ManyToOne
    @JsonIgnore
    @JoinColumn(name = "server_id", insertable = false, updatable = false)
    private UtmServer server;

    @OneToMany(mappedBy = "module", fetch = FetchType.LAZY)
    private Set<UtmModuleGroup> moduleGroups = new HashSet<>();

    @OneToMany(mappedBy = "module", fetch = FetchType.LAZY)
    private Set<UtmLogstashFilter> filters;

    public Boolean getActivatable() {
        return isActivatable;
    }

    public void setActivatable(Boolean activatable) {
        isActivatable = activatable;
    }
}

