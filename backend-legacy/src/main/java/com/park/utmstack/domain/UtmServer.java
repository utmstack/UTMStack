package com.park.utmstack.domain;


import com.fasterxml.jackson.annotation.JsonManagedReference;
import com.park.utmstack.domain.application_modules.UtmModule;

import javax.persistence.*;
import java.io.Serializable;
import java.util.HashSet;
import java.util.Set;

/**
 * A UtmServer.
 */
@Entity
@Table(name = "utm_server")
public class UtmServer implements Serializable {

    private static final long serialVersionUID = 1L;

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(name = "server_name")
    private String serverName;

    @Column(name = "server_type")
    private String serverType;

    @JsonManagedReference
    @OneToMany(mappedBy = "server", fetch = FetchType.EAGER)
    private Set<UtmModule> modules = new HashSet<>();

    public Long getId() {
        return id;
    }

    public void setId(Long id) {
        this.id = id;
    }

    public String getServerName() {
        return serverName;
    }

    public UtmServer serverName(String serverName) {
        this.serverName = serverName;
        return this;
    }

    public void setServerName(String serverName) {
        this.serverName = serverName;
    }

    public String getServerType() {
        return serverType;
    }

    public UtmServer serverType(String serverType) {
        this.serverType = serverType;
        return this;
    }

    public void setServerType(String serverType) {
        this.serverType = serverType;
    }

    public Set<UtmModule> getModules() {
        return modules;
    }

    public void setModules(Set<UtmModule> modules) {
        this.modules = modules;
    }
}
